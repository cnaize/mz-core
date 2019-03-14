package daemon

import (
	"encoding/json"
	"github.com/cnaize/mz-common/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Daemon struct {
	config      Config
	settings    Settings
	stopRefresh chan struct{}
}

func New(config Config) *Daemon {
	return &Daemon{
		config:   config,
		settings: DefaultSettings(),
	}
}

func (d *Daemon) Run() error {
	log.Info("MuzeZone Core: running daemon")
	//if err := d.loadSettings(); err != nil {
	//	if err := d.saveSettings(); err != nil {
	//		log.Warn("Daemon: settings save failed: %+v", err)
	//	}
	//}

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
