package models

// Parameter represents a TR-069 parameter template
type Parameter struct {
	Path              string      `json:"path"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	Type              string      `json:"type"` // string/integer/boolean/unsignedInt/base64/dateTime
	Category          string      `json:"category"`
	Subcategory       string      `json:"subcategory"`
	Writable          bool        `json:"writable"`
	DefaultValue      interface{} `json:"default_value"`
	SupportedVendors  JSONStrings `json:"supported_vendors"`
	SupportedModels   JSONStrings `json:"supported_models"`
	IsStandardTR069   bool        `json:"is_standard_tr069"`
}
