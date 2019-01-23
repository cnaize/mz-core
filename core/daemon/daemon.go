package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
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
	log.Info("MuzeZone Core: running daemon with version: %s", d.config.Version)
	//if err := d.loadSettings(); err != nil {
	//	if err := d.saveSettings(); err != nil {
	//		log.Warn("Daemon: settings save failed: %+v", err)
	//	}
	//}

	return nil
}

func (d *Daemon) RefreshMediaDB() (*model.MediaRootList, error) {
	db := d.config.DB

	if d.stopRefresh != nil {
		close(d.stopRefresh)
		d.stopRefresh = nil
	}

	if err := db.RemoveAllMedia(); err != nil {
		return nil, fmt.Errorf("all media remove failed: %+v", err)
	}

	rootList, err := db.GetMediaRootList()
	if err != nil {
		return nil, fmt.Errorf("media root list get failed: %+v", err)
	}

	d.stopRefresh = make(chan struct{})
	stop := d.stopRefresh

	walkFn := func(path string, info os.FileInfo, err error) error {
		select {
		case <-stop:
			return fmt.Errorf("walk %s stopped", path)
		default:
			if err != nil {
				log.Warn("Daemon: path %s: walk failed: %+v", path, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}

			dir, name, ext := ParseMediaPath(path)
			if ext == model.MediaExtUnknown {
				return nil
			}

			var root *model.MediaRoot
			for _, r := range rootList.Items {
				if dir == r.Dir {
					root = &r
					break
				}
			}
			if root == nil {
				log.Warn("Daemon: path %s: out of roots", path)
				return nil
			}

			media := model.Media{
				Ext:         ext,
				Name:        name,
				Path:        path,
				MediaRootID: root.ID,
			}

			if err := db.AddMedia(&media); err != nil {
				log.Warn("Daemon: media %s add failed: %+v", err)
				return nil
			}
		}

		return nil
	}

	for _, root := range rootList.Items {
		go filepath.Walk(root.Dir, walkFn)
	}

	return rootList, nil
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
