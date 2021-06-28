package manager

import (
	"context"

	"github.com/pkg/errors"
	complianceDatastore "github.com/stackrox/rox/central/compliance/datastore"
	"github.com/stackrox/rox/central/compliance/framework"
	"github.com/stackrox/rox/central/compliance/standards"
	"github.com/stackrox/rox/central/compliance/standards/metadata"
	checkResultsDatastore "github.com/stackrox/rox/central/complianceoperator/checkresults/datastore"
	profileDatastore "github.com/stackrox/rox/central/complianceoperator/profiles/datastore"
	rulesDatastore "github.com/stackrox/rox/central/complianceoperator/rules/datastore"
	scanSettingBindingDatastore "github.com/stackrox/rox/central/complianceoperator/scansettingbinding/datastore"
	"github.com/stackrox/rox/generated/storage"
	pkgFramework "github.com/stackrox/rox/pkg/compliance/framework"
	"github.com/stackrox/rox/pkg/complianceoperator/api/v1alpha1"
	"github.com/stackrox/rox/pkg/logging"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/set"
	"github.com/stackrox/rox/pkg/sync"
)

var (
	log = logging.LoggerForModule()

	allAccessCtx = sac.WithAllAccess(context.Background())

	// errConditionMet is used to short-circuit a walk in the database
	errConditionMet = errors.New("condition met")
)

// Manager helps manage the dynamic profiles from the compliance operator
type Manager interface {
	AddProfile(profile *storage.ComplianceOperatorProfile) error
	DeleteProfile(profile *storage.ComplianceOperatorProfile) error

	AddRule(rule *storage.ComplianceOperatorRule) error
	DeleteRule(rule *storage.ComplianceOperatorRule) error

	IsStandardActive(standardID string) bool
	IsStandardActiveForCluster(standardID, clusterID string) bool

	GetMachineConfigs(clusterID string) ([]string, error)
}

type managerImpl struct {
	registry     *standards.Registry
	registryLock sync.RWMutex

	compliance          complianceDatastore.DataStore
	profiles            profileDatastore.DataStore
	scanSettingBindings scanSettingBindingDatastore.DataStore
	rules               rulesDatastore.DataStore
	results             checkResultsDatastore.DataStore
}

// NewManager returns a new manager of compliance operator resources
func NewManager(registry *standards.Registry, profiles profileDatastore.DataStore, scanSettingBindings scanSettingBindingDatastore.DataStore, rules rulesDatastore.DataStore, results checkResultsDatastore.DataStore, compliance complianceDatastore.DataStore) (Manager, error) {
	mgr := &managerImpl{
		registry: registry,

		compliance:          compliance,
		profiles:            profiles,
		scanSettingBindings: scanSettingBindings,
		rules:               rules,
		results:             results,
	}
	err := profiles.Walk(allAccessCtx, func(profile *storage.ComplianceOperatorProfile) error {
		return mgr.addProfileNoLock(profile)
	})
	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func productTypeToTarget(s string) pkgFramework.TargetKind {
	switch v1alpha1.ComplianceScanType(s) {
	case v1alpha1.ScanTypeNode:
		return pkgFramework.MachineConfigKind
	case v1alpha1.ScanTypePlatform:
		return pkgFramework.ClusterKind
	default:
		return pkgFramework.ClusterKind
	}
}

func getRuleName(rule *storage.ComplianceOperatorRule) string {
	if ruleName, ok := rule.Annotations[v1alpha1.RuleIDAnnotationKey]; ok {
		return ruleName
	}
	// This field is checked within the pipeline so it should never be empty
	log.Errorf("UNEXPECTED: Unknown base rule for %s", rule)
	return "<unknown>"
}

func createControlFromRule(rule *storage.ComplianceOperatorRule) metadata.Control {
	ruleName := getRuleName(rule)
	return metadata.Control{
		ID:          ruleName,
		Name:        ruleName,
		Description: rule.GetTitle(),
	}
}

func (m *managerImpl) createControls(rules []string) ([]metadata.Control, error) {
	controls := make([]metadata.Control, 0, len(rules))
	for _, rule := range rules {
		fullRule, err := m.getRule(rule)
		if err != nil {
			return nil, err
		}
		if fullRule == nil {
			continue
		}
		controls = append(controls, createControlFromRule(fullRule))
	}
	return controls, nil
}

func (m *managerImpl) registerCheckFromRule(standardID string, productType pkgFramework.TargetKind, rule *storage.ComplianceOperatorRule) error {
	ruleName := getRuleName(rule)
	checkMetadata := framework.CheckMetadata{
		ID:                 standards.BuildQualifiedID(standardID, ruleName),
		Scope:              productType,
		InterpretationText: rule.GetDescription(),
	}

	checkFunc := platformCheckFunc(ruleName)
	if productType == pkgFramework.MachineConfigKind {
		checkFunc = machineConfigCheckFunc(ruleName)
	}

	if err := m.registry.RegisterCheck(framework.NewCheckFromFunc(checkMetadata, checkFunc)); err != nil {
		return err
	}
	return nil
}

func (m *managerImpl) AddProfile(profile *storage.ComplianceOperatorProfile) error {
	if err := m.profiles.Upsert(allAccessCtx, profile); err != nil {
		return err
	}

	m.registryLock.Lock()
	defer m.registryLock.Unlock()

	return m.addProfileNoLock(profile)
}

func (m *managerImpl) addProfileNoLock(profile *storage.ComplianceOperatorProfile) error {
	existingProfiles := []*storage.ComplianceOperatorProfile{
		profile,
	}
	if err := m.profiles.Walk(allAccessCtx, func(existingProfile *storage.ComplianceOperatorProfile) error {
		if existingProfile.GetClusterId() != profile.GetClusterId() && existingProfile.GetName() == profile.GetName() {
			existingProfiles = append(existingProfiles, existingProfile)
		}
		return nil
	}); err != nil {
		return err
	}

	standard := metadata.Standard{
		ID:          profile.GetName(),
		Name:        profile.GetName(),
		Description: profile.GetDescription(),
		Dynamic:     true,
	}
	category := metadata.Category{
		ID:          "All",
		Name:        "All",
		Description: "All checks for the profile defined by the Compliance Operator",
	}

	rules := set.NewStringSet()
	for _, profile := range existingProfiles {
		for _, r := range profile.GetRules() {
			rules.Add(r.GetName())
		}
	}
	ruleSlice := rules.AsSortedSlice(func(i, j string) bool {
		return i < j
	})

	var err error
	category.Controls, err = m.createControls(ruleSlice)
	if err != nil {
		return err
	}

	// Get existing standard if it exists, and diff the controls that exist against the current controls
	existingStandard, exists, err := m.registry.Standard(profile.GetName())
	if err != nil {
		return err
	}
	if exists {
		existingControls := set.NewStringSet()
		for _, control := range existingStandard.GetControls() {
			existingControls.Add(control.GetId())
		}
		currentControls := set.NewStringSet()
		for _, control := range category.Controls {
			currentControls.Add(standards.BuildQualifiedID(profile.GetName(), control.ID))
		}
		for controlToDelete := range existingControls.Difference(currentControls) {
			if err := m.registry.DeleteControl(controlToDelete); err != nil {
				return err
			}
		}
	}

	standard.Categories = []metadata.Category{category}

	profileProductType := productTypeToTarget(profile.Annotations[v1alpha1.ProductTypeAnnotation])
	for _, rule := range ruleSlice {
		fullRule, err := m.getRule(rule)
		if err != nil {
			return err
		}
		if fullRule == nil {
			continue
		}

		if err := m.registerCheckFromRule(standard.ID, profileProductType, fullRule); err != nil {
			return errors.Wrapf(err, "registering check %s", fullRule.GetName())
		}
	}

	if err := m.registry.RegisterStandard(standard, true); err != nil {
		log.Errorf("could not register standard %s: %v", profile.GetName(), err)
	}

	return nil
}

func (m *managerImpl) DeleteProfile(deletedProfile *storage.ComplianceOperatorProfile) error {
	if err := m.profiles.Delete(allAccessCtx, deletedProfile.GetId()); err != nil {
		return err
	}

	// ClearAggregationResults when removing a profile as we need to remove cached references
	// to standards that will not be filtered out on the next aggregation call
	if err := m.compliance.ClearAggregationResults(allAccessCtx); err != nil {
		return err
	}

	// Deleting a profile is fairly involved because it involves making sure that the profile name is not referenced
	// anywhere else as standards are indexed by name-based IDs
	m.registryLock.Lock()
	defer m.registryLock.Unlock()

	var found bool
	rulesFound := set.NewStringSet()
	err := m.profiles.Walk(allAccessCtx, func(profile *storage.ComplianceOperatorProfile) error {
		if deletedProfile.GetId() != profile.GetId() && deletedProfile.GetName() == profile.GetName() {
			found = true
			for _, rule := range profile.GetRules() {
				rulesFound.Add(rule.GetName())
			}
		}
		return nil
	})
	if err != nil && err != errConditionMet {
		return err
	}
	if !found {
		if err := m.registry.DeleteStandard(deletedProfile.GetName()); err != nil {
			return err
		}
	}
	for _, rule := range deletedProfile.GetRules() {
		if !rulesFound.Contains(rule.GetName()) {
			rule, err := m.getRule(rule.GetName())
			if err != nil {
				return err
			}
			if rule == nil {
				continue
			}
			if err := m.registry.DeleteControl(standards.BuildQualifiedID(deletedProfile.GetName(), getRuleName(rule))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *managerImpl) IsStandardActive(standardID string) bool {
	standard, ok, err := m.registry.Standard(standardID)
	if err != nil {
		log.Errorf("error getting standard ID %s: %v", standardID, err)
		return false
	}
	if !ok {
		return false
	}
	if !standard.GetMetadata().GetDynamic() {
		return true
	}

	var found bool
	if err := m.scanSettingBindings.Walk(allAccessCtx, func(binding *storage.ComplianceOperatorScanSettingBinding) error {
		for _, p := range binding.GetProfiles() {
			if standardID == p.GetName() {
				found = true
				return errConditionMet
			}
		}
		return nil
	}); err != nil && err != errConditionMet {
		log.Errorf("error walking scan setting bindings datastore: %v", err)
		return false
	}
	return found
}

func (m *managerImpl) IsStandardActiveForCluster(standardID, clusterID string) bool {
	standard, ok, err := m.registry.Standard(standardID)
	if err != nil {
		log.Errorf("error getting standard ID %s: %v", standardID, err)
		return false
	}
	if !ok {
		return false
	}
	if !standard.GetMetadata().GetDynamic() {
		return true
	}

	var found bool
	if err := m.scanSettingBindings.Walk(allAccessCtx, func(binding *storage.ComplianceOperatorScanSettingBinding) error {
		if binding.GetClusterId() == clusterID {
			for _, p := range binding.GetProfiles() {
				if standardID == p.GetName() {
					found = true
					return errConditionMet
				}
			}
		}
		return nil
	}); err != nil && err != errConditionMet {
		log.Errorf("error walking scan setting bindings datastore: %v", err)
		return false
	}
	return found
}

func (m *managerImpl) getRule(name string) (*storage.ComplianceOperatorRule, error) {
	rules, err := m.rules.GetRulesByName(allAccessCtx, name)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}
	return rules[0], nil
}

func (m *managerImpl) GetMachineConfigs(clusterID string) ([]string, error) {
	machineConfigRuleSet := set.NewStringSet()
	err := m.profiles.Walk(allAccessCtx, func(profile *storage.ComplianceOperatorProfile) error {
		if profile.GetClusterId() == clusterID && profile.Annotations[v1alpha1.ProductTypeAnnotation] == string(v1alpha1.ScanTypeNode) {
			for _, profileRule := range profile.GetRules() {
				if rule, err := m.getRule(profileRule.GetName()); err != nil {
					return err
				} else if rule != nil {
					machineConfigRuleSet.Add(getRuleName(rule))
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	machineConfigs := set.NewStringSet()
	err = m.results.Walk(allAccessCtx, func(result *storage.ComplianceOperatorCheckResult) error {
		if result.GetClusterId() != clusterID {
			return nil
		}
		if machineConfigRuleSet.Contains(result.Annotations[v1alpha1.RuleIDAnnotationKey]) {
			if label, ok := result.Labels[v1alpha1.ComplianceScanLabel]; ok {
				machineConfigs.Add(label)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return machineConfigs.AsSlice(), nil
}

func (m *managerImpl) findProfilesWithRuleNoLock(ruleName string) ([]*storage.ComplianceOperatorProfile, error) {
	var profiles []*storage.ComplianceOperatorProfile
	err := m.profiles.Walk(allAccessCtx, func(profile *storage.ComplianceOperatorProfile) error {
		for _, rule := range profile.GetRules() {
			if rule.GetName() == ruleName {
				profiles = append(profiles, profile)
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (m *managerImpl) reindexProfilesWithRuleNoLock(rule *storage.ComplianceOperatorRule) error {
	profiles, err := m.findProfilesWithRuleNoLock(rule.GetName())
	if err != nil {
		return err
	}

	alreadyUpdated := set.NewStringSet()
	for _, profile := range profiles {
		if alreadyUpdated.Add(profile.GetName()) {
			if err := m.addProfileNoLock(profile); err != nil {
				log.Errorf("error updating profile %s: %v", profile.GetName(), err)
			}
		}
	}
	return nil
}

func (m *managerImpl) AddRule(rule *storage.ComplianceOperatorRule) error {
	exists, err := m.rules.ExistsByName(allAccessCtx, rule.GetName())
	if err != nil {
		return err
	}

	if err := m.rules.Upsert(allAccessCtx, rule); err != nil {
		return err
	}
	// No need to reindex if the rule already exists
	if exists {
		return nil
	}

	m.registryLock.Lock()
	defer m.registryLock.Unlock()

	return m.reindexProfilesWithRuleNoLock(rule)
}

func (m *managerImpl) DeleteRule(rule *storage.ComplianceOperatorRule) error {
	if err := m.rules.Delete(allAccessCtx, rule.GetId()); err != nil {
		return err
	}

	m.registryLock.Lock()
	defer m.registryLock.Unlock()

	return m.reindexProfilesWithRuleNoLock(rule)
}