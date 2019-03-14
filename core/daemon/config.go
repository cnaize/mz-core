package daemon

import (
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-core/db"
)

type Config struct {
	WorkingDir    string
	SettingsFile  string
	DatabaseFile  string
	CenterBaseURL string
	CurrentUser   model.User
	DB            db.DB
}

type Settings struct {
	MediaRootList model.MediaRootList `json:"mediaRootList"`
}

func DefaultSettings() Settings {
	return Settings{}
}
