// Code generated by blevebindings generator. DO NOT EDIT.

package mappings

import (
	v1 "github.com/stackrox/rox/generated/api/v1"
	storage "github.com/stackrox/rox/generated/storage"
	search "github.com/stackrox/rox/pkg/search"
)

var OptionsMap = search.Walk(v1.SearchCategory_IMAGE_INTEGRATIONS, "image_integration", (*storage.ImageIntegration)(nil))
