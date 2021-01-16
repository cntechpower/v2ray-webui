package persist

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"
)

var DB *gorm.DB

var ErrDBNotInit = fmt.Errorf("sqlite is not init")

func Init() error {
	header := log.NewHeader("persist.Init")
	var err error
	DB, err = gorm.Open(sqlite.Open("v2ray-webui.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	if err := DB.AutoMigrate(model.GetAllModels()...); err != nil {
		return err
	}
	log.Infof(header, "init sqlite success")

	return nil
}

func Get(m model.Modeler) error {
	return DB.Find(m).Error
}

func Create(m model.Modeler) error {
	if DB == nil {
		return ErrDBNotInit
	}
	return DB.Create(m).Error
}

func Delete(m model.Modeler) error {
	if DB == nil {
		return ErrDBNotInit
	}
	return DB.Delete(m).Error
}

func Save(m model.Modeler) error {
	if DB == nil {
		return ErrDBNotInit
	}
	return DB.Save(m).Error

}
