package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppSettings struct {
	Database DBSettings
	Port     uint16
}

type DBSettings struct {
	User string
	Pass string
	Host string
	Port uint16
	Name string
}

func (db DBSettings) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		db.User, db.Pass, db.Host, db.Port, db.Name,
	)
}

func ConfigureApp() (AppSettings, error) {
	viper.SetConfigFile("./api/configs/dev.yaml")
	if e := viper.ReadInConfig(); e != nil {
		return AppSettings{}, e
	}

	database := DBSettings{
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetUint16("database.port"),
		viper.GetString("database.database_name"),
	}

	port := viper.GetUint16("application_port")

	app := AppSettings{
		Database: database,
		Port:     port,
	}

	return app, nil
}
