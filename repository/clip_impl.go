package repository

import (
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/traPtitech/traQ/model"
)

func (repo *GormRepository) CreateClipFolder(userID uuid.UUID, name string, description string) (*model.ClipFolder, error) {
	uid := uuid.Must(uuid.NewV4())
	//description のバリデーションする
	clipFolder := &model.ClipFolder{
		ID:          uid,
		Description: description,
		Name:        name,
		OwnerID:     userID,
	}
	if err := repo.db.Create(clipFolder).Error; err != nil {
		return nil, err
	}
	return clipFolder, nil
}

func (repo *GormRepository) UpdateClipFolder(folderID uuid.UUID, name string, description string) error {
	if folderID == uuid.Nil {
		return ErrNilID
	}

	var (
		old model.ClipFolder
	)

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&old, &model.ClipFolder{ID: folderID}).Error; err != nil {
			return convertError(err)
		}
		changes := map[string]interface{}{}
		//バリデーションする
		changes["description"] = description
		changes["name"] = name

		// update
		if len(changes) > 0 {
			if err := tx.Model(&old).Updates(changes).Error; err != nil {
				return err
			}
			// updated = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
