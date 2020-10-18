package persist

import (
	"context"
	"time"

	"cntechpower.com/api-server/model"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var cache *redis.Client

func Init(mysqlDsn, redisDsn string) error {
	var err error
	db, err = gorm.Open(mysql.Open(mysqlDsn), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(model.GetAllModels()...); err != nil {
		return err
	}
	cache = redis.NewClient(&redis.Options{
		Addr:     redisDsn,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return nil
}

func MySQL() *gorm.DB {
	return db
}

func Redis() *redis.Client {
	return cache
}

func Get(m model.Modeler) error {
	if err := cache.Get(context.Background(), m.GetCacheKey()).Scan(m); err == nil {
		return nil
	}
	if err := db.Find(m).Error; err != nil {
		return err
	}
	cache.Set(context.Background(), m.GetCacheKey(), m, time.Minute)
	return nil
}

func Create(m model.Modeler) error {
	if err := db.Create(m).Error; err != nil {
		return err
	}
	cache.Set(context.Background(), m.GetCacheKey(), m, time.Minute)
	return nil
}

func Delete(m model.Modeler) error {
	if err := db.Delete(m).Error; err != nil {
		return err
	}
	cache.Del(context.Background(), m.GetCacheKey())
	return nil
}

func BatchGet(ms []model.Modeler) error {
	backToDBModelers := make([]model.Modeler, 0)
	for _, m := range ms {
		if err := cache.Get(context.Background(), m.GetCacheKey()).Scan(m); err != nil {
			backToDBModelers = append(backToDBModelers, m)
		}
	}
	if len(backToDBModelers) == 0 {
		return nil
	}
	for _, backToDBModeler := range backToDBModelers {
		if err := db.Find(backToDBModeler).Error; err != nil {
			return err
		}
		cache.Set(context.Background(), backToDBModeler.GetCacheKey(), backToDBModeler, time.Minute)
	}
	return nil
}
