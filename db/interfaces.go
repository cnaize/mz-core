package db

import "github.com/cnaize/mz-common/model"

type DB interface {
	MediaProvider
}

type MediaProvider interface {
	AddMedia(media *model.Media) error
	RemoveAllMedia() error
	SearchMedia(text string) (*model.MediaList, error)

	GetMediaRootList() (*model.MediaRootList, error)
	AddMediaRoot(root *model.MediaRoot) error
	RemoveMediaRoot(root *model.MediaRoot) error
}
