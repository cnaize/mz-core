package daemon

import (
	"github.com/cnaize/mz-common/model"
	"path/filepath"
)

func ParseMediaPath(path string) (string, string, model.MediaExt) {
	dir, name := filepath.Split(path)

	ext := model.MediaExtUnknown
	for _, smt := range model.SupportedMediaTypes {
		if filepath.Ext(name) == string(smt) {
			ext = smt
			break
		}
	}

	name = name[:len(name)-len(ext)]

	return dir, name, ext
}
