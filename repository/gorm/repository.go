package gorm

import (
	"fmt"

	"github.com/leandro-lugaresi/hub"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/traPtitech/traQ/migration"
	"github.com/traPtitech/traQ/repository"
)

// Repository リポジトリ実装
type Repository struct {
	db     *gorm.DB
	hub    *hub.Hub
	logger *zap.Logger
	stamps *stampRepository
}

// NewGormRepository リポジトリ実装を初期化して生成します。
// スキーマが初期化された場合、init: true を返します。
func NewGormRepository(db *gorm.DB, hub *hub.Hub, logger *zap.Logger, doMigration bool) (repo repository.Repository, init bool, err error) {
	repo = &Repository{
		db:     db,
		hub:    hub,
		logger: logger.Named("repository"),
		stamps: makeStampRepository(db),
	}
	fmt.Println("before migration")
	if doMigration {
		fmt.Println("do migration")
		if init, err = migration.Migrate(db); err != nil {
			fmt.Println("migration error:", err)
			return nil, false, err
		}
		fmt.Println("migration done")
		fmt.Println(init)
	}
	return
}
