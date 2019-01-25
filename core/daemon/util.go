package daemon

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"path/filepath"
)

func ParseMediaPath(path string) (string, string, model.MediaExt, error) {
	dir, name := filepath.Split(path)

	ext := model.MediaExtUnknown
	for _, smt := range model.SupportedMediaTypes {
		if filepath.Ext(name) == string(smt) {
			ext = smt
			break
		}
	}

	if ext == model.MediaExtUnknown {
		return "", "", ext, fmt.Errorf("Daemon: media extension %s doesn't supported yet", filepath.Ext(name))
	}

	name = name[:len(name)-len(ext)]

	return dir, name, ext, nil
}
