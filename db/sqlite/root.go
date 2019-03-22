package sqlite

import "github.com/cnaize/mz-common/model"

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
