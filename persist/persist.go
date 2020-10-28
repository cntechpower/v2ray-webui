package persist

import (
	"context"
	"fmt"
	"time"

	"cntechpower.com/api-server/log"

	"cntechpower.com/api-server/model"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var cache *redis.Client

var ErrDBNotInit = fmt.Errorf("mysql is not init")

func Init(mysqlDsn, redisDsn string) error {
	header := log.NewHeader("persist.Init")
	var err error
	db, err = gorm.Open(mysql.Open(mysqlDsn), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(model.GetAllModels()...); err != nil {
		return err
	}
	log.Infof(header, "init mysql success")
	if redisDsn != "" {
		cache = redis.NewClient(&redis.Options{
			Addr:     redisDsn,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		if err := cache.Ping(context.Background()).Err(); err != nil {
			return err
		}
		log.Infof(header, "init redis success")
	}
	return nil
}

func MySQL() *gorm.DB {
	return db
}

func Redis() *redis.Client {
	return cache
}

func Get(m model.Modeler) error {
	if cache != nil {
		if err := cache.Get(context.Background(), m.GetCacheKey()).Scan(m); err == nil {
			return nil
		}
	}
	if err := db.Find(m).Error; err != nil {
		return err
	}
	if cache != nil {
		cache.Set(context.Background(), m.GetCacheKey(), m, m.GetCacheDuration())
	}
	return nil
}

func Create(m model.Modeler) error {
	if db == nil {
		return ErrDBNotInit
	}
	if err := db.Create(m).Error; err != nil {
		return err
	}
	if cache != nil {
		cache.Set(context.Background(), m.GetCacheKey(), m, m.GetCacheDuration())
	}
	return nil
}

func Delete(m model.Modeler) error {
	if db == nil {
		return ErrDBNotInit
	}
	if err := db.Delete(m).Error; err != nil {
		return err
	}
	if cache != nil {
		cache.Del(context.Background(), m.GetCacheKey())
	}
	return nil
}

func Save(m model.Modeler) error {
	if db == nil {
		return ErrDBNotInit
	}
	if err := db.Save(m).Error; err != nil {
		return err
	}
	if cache != nil {
		cache.Set(context.Background(), m.GetCacheKey(), m, m.GetCacheDuration())
	}
	return nil
}

func BatchGet(ms []model.Modeler) error {
	backToDBModelers := make([]model.Modeler, 0)
	if cache != nil {
		for _, m := range ms {
			if err := cache.Get(context.Background(), m.GetCacheKey()).Scan(m); err != nil {
				backToDBModelers = append(backToDBModelers, m)
			}
		}
	}

	if cache != nil && len(backToDBModelers) == 0 {
		return nil
	}
	for _, backToDBModeler := range backToDBModelers {
		if err := db.Find(backToDBModeler).Error; err != nil {
			return err
		}
		if cache != nil {
			cache.Set(context.Background(), backToDBModeler.GetCacheKey(), backToDBModeler, time.Minute)
		}
	}
	return nil
}
