package sqlite

import (
	"fmt"
	"github.com/cnaize/mz-common/model"
	"strings"
)

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

func (db *DB) SearchMedia(self model.User, request model.SearchRequest, offset, count uint) (model.MediaList, error) {
	db.Lock()
	defer db.Unlock()

	var res model.MediaList
	searchFields := strings.Fields(request.RawText)
	if len(searchFields) < 1 {
		return res, nil
	}

	// handle "..., @username, ..., @username, ..."
	var users []string
	var words []string
	for _, f := range searchFields {
		if strings.HasPrefix(f, "@") {
			users = append(users, f[1:])
		} else {
			words = append(words, f)
		}
	}

	if len(users) > 0 {
		var selfFound bool
		for _, u := range users {
			if u == strings.ToLower(self.Username) {
				selfFound = true
				break
			}
		}

		if !selfFound {
			return res, nil
		}
	}


	search := fmt.Sprintf("%%%s%%", strings.Join(words, "%"))
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
