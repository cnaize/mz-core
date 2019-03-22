package daemon

import (
	"encoding/json"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"github.com/cnaize/mz-core/db"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Daemon struct {
	DB               db.DB
	CurrentUser      model.User
	config           Config
	baseReq          *gorequest.SuperAgent
	runOnce          sync.Once
	searchReqOffset  uint
	stopMediaRefresh chan struct{}
	settings         Settings
}

func New(config Config, db db.DB) *Daemon {
	return &Daemon{
		DB:       db,
		config:   config,
		settings: DefaultSettings(),
	}
}

func (d *Daemon) Run() error {
	log.Info("MuzeZone Core: running daemon")

	d.baseReq = gorequest.New().Timeout(500 * time.Millisecond)

	//if err := d.loadSettings(); err != nil {
	//	if err := d.saveSettings(); err != nil {
	//		log.Warn("Daemon: settings save failed: %+v", err)
	//	}
	//}

	return nil
}

func (d *Daemon) StartFeedLoop(user model.User) error {
	d.CurrentUser = user
	d.baseReq.Set("Authorization", "Bearer "+user.Token)

	d.runOnce.Do(func() {
		go d.searchFeedLoop()
		go d.mediaFeedLoop()
	})

	return nil
}

func (d *Daemon) loadSettings() error {
	settingsPath := filepath.Join(d.config.WorkingDir, d.config.SettingsFile)
	settingsData, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(settingsData, &d.settings); err != nil {
		return err
	}

	log.Info("Daemon: settings are loaded")
	return nil
}

func (d *Daemon) saveSettings() error {
	settingsPath := filepath.Join(d.config.WorkingDir, d.config.SettingsFile)
	settingsData, err := json.Marshal(d.settings)
	if err != nil {
		return err
	}

	if _, err := os.Stat(d.config.WorkingDir); err != nil {
		if err := os.MkdirAll(d.config.WorkingDir, 0755); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(settingsPath, settingsData, 0755); err != nil {
		return err
	}

	log.Info("Daemon: settings are created")
	return nil
}
