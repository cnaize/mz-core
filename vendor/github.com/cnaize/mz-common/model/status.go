package model

type CenterStatus struct {
	MinCoreVersion *string `json:"minCoreVersion,omitempty" form:"minCoreVersion"`
}

type CoreStatus struct {
	Version *string `json:"version,omitempty" form:"version"`
}
