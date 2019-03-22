package db

import "github.com/cnaize/mz-common/model"

type DB interface {
	MediaProvider
}

type MediaProvider interface {
	GetMediaByID(id uint) (model.Media, error)
	AddMedia(media model.Media) error

	SearchMedia(request model.SearchRequest, offset, count uint) (model.MediaList, error)
	RemoveAllMedia() error

	GetMediaRootList() (model.MediaRootList, error)
	AddMediaRoot(root model.MediaRoot) error
	RemoveMediaRoot(root model.MediaRoot) error

	IsMediaItemNotFound(err error) bool
}
