package ioc

import (
	"webook/config"
	dao "webook/internal/adapters/outbound/persistence/mysql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewDB 创建数据库连接
func NewDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 自动迁移数据库表结构
	err = db.AutoMigrate(
		&dao.User{},
		&dao.Post{},
		&dao.PublishedPost{},
		&dao.PostStats{},
		&dao.PostLikeRelation{},
		&dao.PostCollectRelation{},
	)
	if err != nil {
		panic(err)
	}
	return db
}
