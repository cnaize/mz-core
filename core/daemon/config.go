package daemon

import (
	"github.com/cnaize/mz-common/model"
)

type Config struct {
	WorkingDir    string
	SettingsFile  string
	DatabaseFile  string
	CenterBaseURL string
}

type Settings struct {
	MediaRootList model.MediaRootList `json:"mediaRootList"`
}

func DefaultSettings() Settings {
	return Settings{}
}
