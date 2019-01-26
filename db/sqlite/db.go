package sqlite

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"strings"
)

type DB struct {
	db *gorm.DB
}

func New(filepath string) (*DB, error) {
	db, err := gorm.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("open failed: %+v", err)
	}

	if err := prepare(db); err != nil {
		return nil, fmt.Errorf("prepare failed: %+v", err)
	}

	return &DB{
		db: db,
	}, nil
}

func (db *DB) AddMedia(media model.Media) error {
	return db.db.Create(&media).Error
}

func (db *DB) SearchMedia(request model.SearchRequest) (model.MediaList, error) {
	search := fmt.Sprintf("%%%s%%", strings.Join(strings.Split(request.RawText, " "), "%"))

	var res model.MediaList
	// TODO: change hardcoded "access_type" after sign up implementation
	if err := db.db.Joins("INNER JOIN media_roots ON media_roots.id = media.media_root_id").
		Where("media_roots.access_type = ?", model.MediaAccessTypePublic).
		Where("media.raw_path LIKE ?", search).Find(&res.Items).Error; err != nil {
		return res, err
	}

	return res, nil
}

func (db *DB) RemoveAllMedia() error {
	if err := db.db.DropTable(&model.Media{}).Error; err != nil {
		return err
	}

	if err := db.db.AutoMigrate(&model.Media{}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) GetMediaRootList() (model.MediaRootList, error) {
	var res model.MediaRootList
	if err := db.db.Find(&res.Items).Error; err != nil {
		return res, err
	}

	for _, r := range res.Items {
		db.db.Model(&model.Media{}).Where("media_root_id = ?", r.ID).Count(&r.MediaCount)
	}

	return res, nil
}

func (db *DB) AddMediaRoot(root model.MediaRoot) error {
	return db.db.Create(&root).Error
}

func (db *DB) RemoveMediaRoot(root model.MediaRoot) error {
	return db.db.Delete(model.MediaRoot{}, "id = ?", root.ID).Error
}

func (db *DB) IsMediaItemNotFound(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}

func prepare(db *gorm.DB) error {
	db.LogMode(true)

	if err := db.AutoMigrate(&model.MediaRoot{}).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Media{}).Error; err != nil {
		return err
	}

	return nil
}
