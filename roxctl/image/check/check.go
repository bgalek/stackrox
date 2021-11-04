package check

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/images/utils"
	"github.com/stackrox/rox/pkg/retry"
	pkgCommon "github.com/stackrox/rox/pkg/roxctl/common"
	pkgUtils "github.com/stackrox/rox/pkg/utils"
	"github.com/stackrox/rox/roxctl/common/environment"
	"github.com/stackrox/rox/roxctl/common/flags"
	"github.com/stackrox/rox/roxctl/common/printer"
	"github.com/stackrox/rox/roxctl/common/report"
	"github.com/stackrox/rox/roxctl/summaries/policy"
)

const (
	jsonFlagName     = "json"
	jsonFailFlagName = "json-fail-on-policy-violations"
)

var (
	// Default headers to use when printing tabular output
	defaultImageCheckHeaders = []string{
		"POLICY", "SEVERITY", "DESCRIPTION", "VIOLATION", "REMEDIATION",
	}
	// Default JSON path expression which retrieves the data from the policyJSONResult
	defaultImageCheckJSONPathExpression = "{results.#.violatedPolicies.#.name," +
		"results.#.violatedPolicies.#.severity," +
		"results.#.violatedPolicies.#.description," +
		"results.#.violatedPolicies.#.violation," +
		"results.#.violatedPolicies.#.remediation}"
	// supported output formats with default values
	supportedObjectPrinters = []printer.CustomPrinterFactory{
		printer.NewTabularPrinterFactory(false, defaultImageCheckHeaders, defaultImageCheckJSONPathExpression, false, false),
		printer.NewJSONPrinterFactory(false, false),
	}
)

// Command checks the image against image build lifecycle policies
func Command(cliEnvironment environment.Environment) *cobra.Command {
	imageCheckCmd := &imageCheckCommand{env: cliEnvironment}

	// object printer factory - allows output formats of JSON, csv, table with table being the default
	objectPrinterFactory, err := printer.NewObjectPrinterFactory("table", supportedObjectPrinters...)
	// the returned error only occurs when default values do not allow the creation of any printer, this should be considered
	// a programming error rather than a user error
	pkgUtils.Must(err)

	c := &cobra.Command{
		Use:  "check",
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if err := imageCheckCmd.Construct(nil, c, objectPrinterFactory); err != nil {
				return err
			}
			if err := imageCheckCmd.Validate(); err != nil {
				return err
			}
			return imageCheckCmd.CheckImage()
		},
	}

	// Add all flags required by the printer factories with the provided default values
	objectPrinterFactory.AddFlags(c)

	// Image Check specific flags
	c.Flags().StringVarP(&imageCheckCmd.image, "image", "i", "", "image name and reference. (e.g. nginx:latest or nginx@sha256:...)")
	pkgUtils.Must(c.MarkFlagRequired("image"))
	c.Flags().IntVarP(&imageCheckCmd.retryDelay, "retry-delay", "d", 3, "set time to wait between retries in seconds.")
	c.Flags().IntVarP(&imageCheckCmd.retryCount, "retries", "r", 0, "number of retries before exiting as error.")
	c.Flags().BoolVar(&imageCheckCmd.sendNotifications, "send-notifications", false,
		"whether to send notifications for violations (notifications will be sent to the notifiers "+
			"configured in each violated policy).")
	c.Flags().StringSliceVarP(&imageCheckCmd.policyCategories, "categories", "c", nil, "optional comma separated list of policy categories to run.  Defaults to all policy categories.")

	// deprecated, old output format specific flags
	c.Flags().BoolVar(&imageCheckCmd.printAllViolations, "print-all-violations", false, "whether to print all violations per alert or truncate violations for readability")
	c.Flags().BoolVar(&imageCheckCmd.json, jsonFlagName, false, "Output policy results as JSON")
	c.Flags().BoolVar(&imageCheckCmd.failViolationsWithJSON, jsonFailFlagName, true,
		"Whether policy violations should cause the command to exit non-zero in JSON output mode too. "+
			"This flag only has effect when --json is also specified.")
	// mark old output format flags as deprecated, but do not fully remove them to not break API for customer
	// each deprecation message will be prefixed with "<flag-name> is deprecated,"
	pkgUtils.Must(c.Flags().MarkDeprecated("print-all-violations", "use the new output format which handles this by default. The flag is only "+
		"relevant in combination with the --json flag"))
	pkgUtils.Must(c.Flags().MarkDeprecated(jsonFlagName, "use the new output format which also offers JSON. NOTE: The new output format's structure "+
		"has changed in a non-backward compatible way."))
	pkgUtils.Must(c.Flags().MarkDeprecated(jsonFailFlagName, "use the new output format which will always fail with policy violations."))

	return c
}

// imageCheckCommand holds all configurations and metadata to execute an image check
type imageCheckCommand struct {
	// properties bound to cobra flags
	image              string
	retryDelay         int
	retryCount         int
	sendNotifications  bool
	policyCategories   []string
	printAllViolations bool
	timeout            time.Duration

	// values injected from either Construct, parent command or for abstracting external dependencies
	env                      environment.Environment
	objectPrinter            printer.ObjectPrinter
	standardizedOutputFormat bool

	// TODO: Remove these values once the old format is fully deprecated
	// values of deprecated flags
	json                   bool
	failViolationsWithJSON bool
}

// Construct will enhance the struct with other values coming either from os.Args, other, global flags or environment variables
func (i *imageCheckCommand) Construct(args []string, cmd *cobra.Command, f *printer.ObjectPrinterFactory) error {
	i.timeout = flags.Timeout(cmd)

	// TODO: remove this once we have fully deprecated the old output format
	// Only create a printer when --json is not given
	if !i.json {
		p, err := f.CreatePrinter()
		if err != nil {
			return err
		}
		i.objectPrinter = p
		i.standardizedOutputFormat = f.IsStandardizedFormat()
		// Silence errors when a standardized output format is used to make sure the output format is not "destroyed"
		// due to an error string, i.e. when policies are failing the check
		cmd.SilenceErrors = i.standardizedOutputFormat
	}

	return nil
}

// Validate will validate the injected values and check whether it's possible to execute the operation with the
// provided values
func (i *imageCheckCommand) Validate() error {

	// TODO: remove this once we have fully deprecated the old output format
	// Only print warnings specific to old --json format when no printer is created
	if i.objectPrinter == nil {
		if i.failViolationsWithJSON && !i.json {
			fmt.Fprintf(i.env.InputOutput().ErrOut, "Note: --%s has no effect when --%s is not specified.\n", jsonFailFlagName, jsonFlagName)
		}
	}

	return nil
}

// CheckImage will execute the image check with retry functionality
func (i *imageCheckCommand) CheckImage() error {
	err := retry.WithRetry(func() error {
		return i.checkImage()
	},
		retry.Tries(i.retryCount+1),
		retry.OnFailedAttempts(func(err error) {
			fmt.Fprintf(i.env.InputOutput().ErrOut, "Checking image failed: %v. Retrying after %v seconds\n", err, i.retryDelay)
			time.Sleep(time.Duration(i.retryDelay) * time.Second)
		}))
	if err != nil {
		return err
	}
	return nil
}

func (i *imageCheckCommand) checkImage() error {
	// Get the violated policies for the input data.
	req, err := buildRequest(i.image, i.sendNotifications, i.policyCategories)
	if err != nil {
		return err
	}
	alerts, err := i.getAlerts(req)
	if err != nil {
		return err
	}
	return i.printResults(alerts)
}

func (i *imageCheckCommand) printResults(alerts []*storage.Alert) error {
	// create the alert summary object
	policySummary := policy.NewPolicySummaryForPrinting(alerts, storage.EnforcementAction_FAIL_BUILD_ENFORCEMENT)
	amountBuildBreakingPolicies := policySummary.GetTotalAmountOfBreakingPolicies()

	// TODO: Remove this once the old output format is fully deprecated
	// Legacy printing based on whether --json is set to true or not.
	if i.json {
		return legacyPrint(alerts, i.failViolationsWithJSON, amountBuildBreakingPolicies, i.env.InputOutput().Out)
	}

	// conditionally print a summary when the output format is a "non-RFC/standardized" one
	// could be -> text, wide, tree etc.
	if !i.standardizedOutputFormat {
		printPolicySummary(i.image, policySummary.Summary, i.env.InputOutput().Out)
	}

	// print the JSON object in the dedicated format via a printer.ObjectPrinter
	if err := i.objectPrinter.Print(policySummary, i.env.InputOutput().Out); err != nil {
		return err
	}

	// conditionally print errors when the output format is a "non-RFC/standardized" one
	// could be -> text, wide, tree etc.
	if !i.standardizedOutputFormat {
		printAdditionalWarnsAndErrs(policySummary.Summary[policy.TotalPolicyAmountKey], policySummary.Results,
			amountBuildBreakingPolicies, i.env.InputOutput().Out)
	}

	if amountBuildBreakingPolicies != 0 {
		return policy.NewErrBreakingPolicies(amountBuildBreakingPolicies)
	}
	return nil
}

func (i *imageCheckCommand) getAlerts(req *v1.BuildDetectionRequest) ([]*storage.Alert, error) {
	conn, err := i.env.GRPCConnection()
	if err != nil {
		return nil, err
	}

	defer pkgUtils.IgnoreError(conn.Close)
	svc := v1.NewDetectionServiceClient(conn)

	ctx, cancel := context.WithTimeout(pkgCommon.Context(), i.timeout)
	defer cancel()

	response, err := svc.DetectBuildTime(ctx, req)
	if err != nil {
		return nil, err
	}

	return response.GetAlerts(), err
}

// legacyPrint supports the old printing behavior of the --json format to ensure backwards compatability
func legacyPrint(alerts []*storage.Alert, failViolations bool, numBuildBreakingPolicies int, out io.Writer) error {
	err := report.JSON(out, alerts)
	if err != nil {
		return err
	}
	if failViolations && numBuildBreakingPolicies != 0 {
		return errors.New("Violated a policy with CI enforcement set")
	}
	return nil
}

// printPolicySummary prints a header with an overview of all found policy violations by policySeverity for
// non-standardized output format, i.e. table format
func printPolicySummary(image string, numOfPolicyViolations map[string]int, out io.Writer) {
	fmt.Fprintf(out, "Policy check results for image: %s\n", image)
	fmt.Fprintf(out, "(%s: %d, %s: %d, %s: %d, %s: %d, %s: %d)\n\n",
		policy.TotalPolicyAmountKey, numOfPolicyViolations[policy.TotalPolicyAmountKey],
		policy.LowSeverity, numOfPolicyViolations[policy.LowSeverity.String()],
		policy.MediumSeverity, numOfPolicyViolations[policy.MediumSeverity.String()],
		policy.HighSeverity, numOfPolicyViolations[policy.HighSeverity.String()],
		policy.CriticalSeverity, numOfPolicyViolations[policy.CriticalSeverity.String()])
}

// printAdditionalWarnsAndErrs prints a warning indicating how many policies have been failed as well as errors for each
// policy that failed the check. This will be printed only for non-standardized output formats, i.e. table format
// and if there are any failed policies
func printAdditionalWarnsAndErrs(numTotalViolatedPolicies int, results []policy.EntityResult, numBreakingPolicies int, out io.Writer) {
	if numTotalViolatedPolicies == 0 {
		return
	}
	fmt.Fprintf(out, "WARN: A total of %d policies have been violated\n", numTotalViolatedPolicies)

	if numBreakingPolicies == 0 {
		return
	}
	fmt.Fprintf(out, "ERROR: %s\n", policy.NewErrBreakingPolicies(numBreakingPolicies))

	for _, res := range results {
		for _, breakingPolicy := range res.BreakingPolicies {
			fmt.Fprintf(out, "ERROR: Policy %q - Possible remediation: %q\n",
				breakingPolicy.Name, breakingPolicy.Remediation)
		}
	}
}

// Use inputs to generate an image name for request.
func buildRequest(image string, sendNotifications bool, policyCategories []string) (*v1.BuildDetectionRequest, error) {
	img, err := utils.GenerateImageFromString(image)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse image '%s'", image)
	}
	return &v1.BuildDetectionRequest{
		Resource:          &v1.BuildDetectionRequest_Image{Image: img},
		SendNotifications: sendNotifications,
		PolicyCategories:  policyCategories,
	}, nil
}
