package sqlite

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"strings"
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

func (db *DB) GetMediaByID(id uint) (model.Media, error) {
	db.Lock()
	defer db.Unlock()

	var res model.Media
	if err := db.db.First(&res, id).Error; err != nil {
		return res, err
	}

	return res, nil
}

func (db *DB) AddMedia(media model.Media) error {
	db.Lock()
	defer db.Unlock()

	return db.db.Create(&media).Error
}

func (db *DB) SearchMedia(request model.SearchRequest, offset, count uint) (model.MediaList, error) {
	db.Lock()
	defer db.Unlock()

	var res model.MediaList
	searchFields := strings.Fields(request.RawText)
	if len(searchFields) < 1 {
		return res, nil
	}
	if strings.HasPrefix(searchFields[0], "@") {
		searchFields[0] = searchFields[0] + "/"
	}

	search := fmt.Sprintf("%%%s%%", strings.Join(searchFields, "%"))
	modes := []model.MediaAccessType{request.Mode}
	// "protected" includes "public, "private" includes "public" and "protected"
	if request.Mode == model.MediaAccessTypeProtected || request.Mode == model.MediaAccessTypePrivate {
		modes = append(modes, model.MediaAccessTypePublic)
	}
	if request.Mode == model.MediaAccessTypePrivate {
		modes = append(modes, model.MediaAccessTypeProtected)
	}

	query := db.db.Joins("INNER JOIN media_roots ON media_roots.id = media.media_root_id").
		Where("media.raw_path LIKE ?", search).
		Where("media_roots.access_type IN (?)", modes)

	if err := query.Offset(offset).Limit(count).Find(&res.Items).Error; err != nil {
		return res, err
	}

	query.Model(&model.Media{}).Count(&res.AllItemsCount)

	return res, nil
}

func (db *DB) RemoveAllMedia() error {
	db.Lock()
	defer db.Unlock()

	if err := db.db.DropTable(&model.Media{}).Error; err != nil {
		return err
	}

	if err := db.db.AutoMigrate(&model.Media{}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) GetMediaRootList() (model.MediaRootList, error) {
	db.Lock()
	defer db.Unlock()

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
	db.Lock()
	defer db.Unlock()

	return db.db.Create(&root).Error
}

func (db *DB) RemoveMediaRoot(root model.MediaRoot) error {
	db.Lock()
	defer db.Unlock()

	return db.db.Delete(&root).Error
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
