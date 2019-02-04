package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func (d *Daemon) RefreshMediaDB() error {
	db := d.config.DB

	if d.stopRefresh != nil {
		close(d.stopRefresh)
		d.stopRefresh = nil
	}

	if err := db.RemoveAllMedia(); err != nil {
		return err
	}

	rootList, err := db.GetMediaRootList()
	if err != nil {
		return err
	}

	d.stopRefresh = make(chan struct{})
	stop := d.stopRefresh

	walk := func(root model.MediaRoot) {
		if err := filepath.Walk(root.Dir, func(path string, info os.FileInfo, err error) error {
			select {
			case <-stop:
				return fmt.Errorf("terminated")
			default:
				if err != nil {
					log.Warn("Daemon: path %s: walk failed: %+v", path, err)
					return nil
				}
				if info.IsDir() {
					return nil
				}

				dir, name, ext, err := parseMediaPath(path)
				if err != nil {
					return nil
				}

				for _, r := range rootList.Items {
					if strings.HasPrefix(dir, r.Dir) && len(r.Dir) > len(root.Dir) {
						return nil
					}
				}

				media := model.Media{
					Name:        name,
					Ext:         ext,
					Dir:         dir[len(root.Dir):],
					RawPath:     strings.ToLower(path[len(root.Dir):]),
					MediaRootID: root.ID,
				}

				if err := db.AddMedia(media); err != nil {
					log.Warn("Daemon: media %s add failed: %+v", err)
					return nil
				}
			}

			return nil
		}); err != nil {
			log.Info("Daemon: path %s: walk stopped: %+v", err)
		}
	}

	for _, root := range rootList.Items {
		go walk(*root)
	}

	return nil
}

func parseMediaPath(path string) (string, string, model.MediaExt, error) {
	dir, name := filepath.Split(path)

	fileExt := filepath.Ext(name)
	if fileExt == "" {
		return "", "", "", fmt.Errorf("file in path %s has empty extension", path)
	}

	ext := model.MediaExtUnknown
	for _, smt := range model.SupportedMediaExtList {
		if fileExt[1:] == string(smt) {
			ext = smt
			break
		}
	}

	if ext == model.MediaExtUnknown {
		return "", "", "", fmt.Errorf("media with extension %s doesn't supported yet", fileExt)
	}

	name = name[:len(name)-len(ext)-1]

	return dir, name, ext, nil
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
