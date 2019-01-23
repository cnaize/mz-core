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

	prepare(db)

	return &DB{
		db: db,
	}, nil
}

func (db *DB) AddMedia(media *model.Media) error {
	return db.db.Create(&media).Error
}

func (db *DB) RemoveAllMedia() error {
	if err := db.db.DropTable(&model.Media{}).Error; err != nil {
		return err
	}

	db.db.AutoMigrate(&model.Media{})

	return nil
}

func (db *DB) SearchMedia(text string) (*model.MediaList, error) {
	search := fmt.Sprintf("%%%s%%", strings.Join(strings.Split(text, " "), "%"))

	var res model.MediaList
	if err := db.db.Where("path LIKE ?", search).Find(&res.Items).Error; err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *DB) GetMediaRootList() (*model.MediaRootList, error) {
	var res model.MediaRootList
	if err := db.db.Find(&res.Items).Error; err != nil {
		return nil, err
	}

	for _, r := range res.Items {
		db.db.Model(&model.Media{}).Where("media_root_id = ?", r.ID).Count(&r.ItemsCount)
	}

	return &res, nil
}

func (db *DB) AddMediaRoot(root *model.MediaRoot) error {
	return db.db.Create(root).Error
}

func (db *DB) RemoveMediaRoot(root *model.MediaRoot) error {
	return db.db.Where("dir = ?", root.Dir).Error
}

func prepare(db *gorm.DB) {
	db.LogMode(true)

	db.AutoMigrate(&model.MediaRoot{})
	db.AutoMigrate(&model.Media{})
}
