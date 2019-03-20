package model

type CenterStatus struct {
	MinCoreVersion string `json:"minCoreVersion" form:"minCoreVersion"`
}

type CoreStatus struct {
	Version string `json:"version" form:"version"`
}
