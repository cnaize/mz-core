package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/log"
	"github.com/cnaize/mz-common/model"
	"os"
	"path/filepath"
	"strings"
)

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
					// TODO: add username as prefix after singup implementation
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
