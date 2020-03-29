package repository

import (
	"encoding/hex"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/leandro-lugaresi/hub"
	"github.com/traPtitech/traQ/event"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/utils"
	"github.com/traPtitech/traQ/utils/validator"
	"unicode/utf8"
)

// CreateUser implements UserRepository interface.
func (repo *GormRepository) CreateUser(args CreateUserArgs) (model.UserInfo, error) {
	uid := uuid.Must(uuid.NewV4())
	user := &model.User{
		ID:          uid,
		Name:        args.Name,
		DisplayName: args.DisplayName,
		Status:      model.UserAccountStatusActive,
		Bot:         false,
		Role:        args.Role,
		Profile:     &model.UserProfile{UserID: uid},
	}

	if len(args.Password) > 0 {
		salt := utils.GenerateSalt()
		user.Password = hex.EncodeToString(utils.HashPassword(args.Password, salt))
		user.Salt = hex.EncodeToString(salt)
	}

	if args.IconFileID.Valid {
		user.Icon = args.IconFileID.UUID
	} else {
		iconID, err := GenerateIconFile(repo, user.Name)
		if err != nil {
			return nil, err
		}
		user.Icon = iconID
	}

	if args.ExternalLogin != nil {
		args.ExternalLogin.UserID = uid
	}

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if exist, err := dbExists(tx, &model.User{Name: user.Name}); err != nil {
			return err
		} else if exist {
			return ErrAlreadyExists
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if err := tx.Create(user.Profile).Error; err != nil {
			return err
		}
		if args.ExternalLogin != nil {
			if err := tx.Create(args.ExternalLogin).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	repo.hub.Publish(hub.Message{
		Name: event.UserCreated,
		Fields: hub.Fields{
			"user_id": user.ID,
			"user":    user,
		},
	})
	return user, nil
}

// GetUser implements UserRepository interface.
func (repo *GormRepository) GetUser(id uuid.UUID, withProfile bool) (model.UserInfo, error) {
	if id == uuid.Nil {
		return nil, ErrNotFound
	}
	return getUser(repo.db, withProfile, &model.User{ID: id})
}

// GetUserByName implements UserRepository interface.
func (repo *GormRepository) GetUserByName(name string, withProfile bool) (model.UserInfo, error) {
	if len(name) == 0 {
		return nil, ErrNotFound
	}
	return getUser(repo.db, withProfile, &model.User{Name: name})
}

func getUser(tx *gorm.DB, withProfile bool, where ...interface{}) (model.UserInfo, error) {
	var user model.User
	if withProfile {
		tx = tx.Preload("Profile")
	}
	if err := tx.First(&user, where...).Error; err != nil {
		return nil, convertError(err)
	}
	return &user, nil
}

// GetUsers implements UserRepository interface.
func (repo *GormRepository) GetUsers(query UsersQuery) (users []model.UserInfo, err error) {
	arr := make([]*model.User, 0)
	if err = repo.makeGetUsersTx(query).Find(&arr).Error; err != nil {
		return nil, err
	}

	users = make([]model.UserInfo, len(arr))
	for i, u := range arr {
		users[i] = u
	}
	return users, nil
}

// GetUserIDs implements UserRepository interface.
func (repo *GormRepository) GetUserIDs(query UsersQuery) (ids []uuid.UUID, err error) {
	ids = make([]uuid.UUID, 0)
	err = repo.makeGetUsersTx(query).Pluck("users.id", &ids).Error
	return ids, err
}

func (repo *GormRepository) makeGetUsersTx(query UsersQuery) *gorm.DB {
	tx := repo.db.Table("users")

	if query.IsActive.Valid {
		if query.IsActive.Bool {
			tx = tx.Where("users.status = ?", model.UserAccountStatusActive)
		} else {
			tx = tx.Where("users.status != ?", model.UserAccountStatusActive)
		}
	}
	if query.IsBot.Valid {
		tx = tx.Where("users.bot = ?", query.IsBot.Bool)
	}
	if query.IsSubscriberAtMarkLevelOf.Valid {
		tx = tx.Joins("INNER JOIN users_subscribe_channels ON users_subscribe_channels.user_id = users.id AND users_subscribe_channels.channel_id = ? AND users_subscribe_channels.mark = true", query.IsSubscriberAtMarkLevelOf.UUID)
	}
	if query.IsSubscriberAtNotifyLevelOf.Valid {
		tx = tx.Joins("INNER JOIN users_subscribe_channels ON users_subscribe_channels.user_id = users.id AND users_subscribe_channels.channel_id = ? AND users_subscribe_channels.notify = true", query.IsSubscriberAtNotifyLevelOf.UUID)
	}
	if query.IsCMemberOf.Valid {
		tx = tx.Joins("INNER JOIN users_private_channels ON users_private_channels.user_id = users.id AND users_private_channels.channel_id = ?", query.IsCMemberOf.UUID)
	}
	if query.IsGMemberOf.Valid {
		tx = tx.Joins("INNER JOIN user_group_members ON user_group_members.user_id = users.id AND user_group_members.group_id = ?", query.IsGMemberOf.UUID)
	}
	if query.EnableProfileLoading {
		tx = tx.Preload("Profile")
	}

	return tx
}

// UserExists implements UserRepository interface.
func (repo *GormRepository) UserExists(id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, nil
	}
	return dbExists(repo.db, &model.User{ID: id})
}

// UpdateUser implements UserRepository interface.
func (repo *GormRepository) UpdateUser(id uuid.UUID, args UpdateUserArgs) error {
	if id == uuid.Nil {
		return ErrNilID
	}
	var (
		changed bool
		count   int
	)
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		var u model.User
		if err := tx.Preload("Profile").First(&u, model.User{ID: id}).Error; err != nil {
			return convertError(err)
		}

		changes := map[string]interface{}{}
		if args.DisplayName.Valid {
			if utf8.RuneCountInString(args.DisplayName.String) > 64 {
				return ArgError("args.DisplayName", "DisplayName must be shorter than 64 characters")
			}
			changes["display_name"] = args.DisplayName.String
		}
		if args.Role.Valid {
			changes["role"] = args.Role.String
		}
		if args.UserState.Valid {
			changes["status"] = args.UserState.State.Int()
		}
		if args.IconFileID.Valid {
			changes["icon"] = args.IconFileID.UUID
		}
		if args.Password.Valid {
			salt := utils.GenerateSalt()
			changes["salt"] = hex.EncodeToString(salt)
			changes["password"] = hex.EncodeToString(utils.HashPassword(args.Password.String, salt))
		}
		if len(changes) > 0 {
			if err := tx.Model(&u).Updates(changes).Error; err != nil {
				return err
			}
			changed = true
			count += len(changes)
		}

		changes = map[string]interface{}{}
		if args.TwitterID.Valid {
			if len(args.TwitterID.String) > 0 && !validator.TwitterIDRegex.MatchString(args.TwitterID.String) {
				return ArgError("args.TwitterID", "invalid TwitterID")
			}
			changes["twitter_id"] = args.TwitterID.String
		}
		if args.Bio.Valid {
			changes["bio"] = args.Bio.String
		}
		if args.LastOnline.Valid {
			changes["last_online"] = args.LastOnline
		}
		if len(changes) > 0 {
			if err := tx.Model(u.Profile).Updates(changes).Error; err != nil {
				return err
			}
			changed = true
			count += len(changes)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if args.Password.Valid && count == 2 {
		return nil // パスワードのみの変更の時はUserUpdatedイベントを発生させない
	}
	if args.LastOnline.Valid && count == 1 {
		return nil // 最終オンライン日時のみの更新の時はUserUpdatedイベントを発生させない
	}
	if changed {
		if args.IconFileID.Valid && count == 1 {
			repo.hub.Publish(hub.Message{
				Name: event.UserIconUpdated,
				Fields: hub.Fields{
					"user_id": id,
					"file_id": args.IconFileID.UUID,
				},
			})
		} else {
			repo.hub.Publish(hub.Message{
				Name: event.UserUpdated,
				Fields: hub.Fields{
					"user_id": id,
				},
			})
		}
	}
	return nil
}
