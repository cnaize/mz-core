package sqlite

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"sync"
)

type DB struct {
	sync.Mutex
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

func (db *DB) IsMediaItemNotFound(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}

func prepare(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.MediaRoot{}).Error; err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Media{}).Error; err != nil {
		return err
	}

	return nil
}
