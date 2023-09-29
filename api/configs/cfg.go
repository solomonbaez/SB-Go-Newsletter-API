package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

const CFG = "./api/configs/dev.yaml"

func init() {
	viper.SetConfigFile(CFG)
}

// APPLICATION
type AppSettings struct {
	Database DBSettings
	Port     uint16
}

type DBSettings struct {
	user string
	pass string
	host string
	port uint16
	name string
}

func (db DBSettings) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		db.user, db.pass, db.host, db.port, db.name,
	)
}

func ConfigureApp() (*AppSettings, error) {
	if e := viper.ReadInConfig(); e != nil {
		return nil, e
	}

	database := DBSettings{
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetUint16("database.port"),
		viper.GetString("database.database_name"),
	}

	port := viper.GetUint16("application_port")

	settings := &AppSettings{
		Database: database,
		Port:     port,
	}

	return settings, nil
}

// EMAIL CLIENT
type EmailClientSettings struct {
	Server   string
	Port     int
	Username string
	Password string
	Sender   string
}

func ConfigureEmailClient(cfg string) (*EmailClientSettings, error) {
	if cfg != "" {
		viper.SetConfigFile(cfg)
	}
	if e := viper.ReadInConfig(); e != nil {
		return nil, e
	}

	settings := &EmailClientSettings{
		viper.GetString("email.server"),
		viper.GetInt("email.port"),
		viper.GetString("email.username"),
		viper.GetString("email.password"),
		viper.GetString("email.sender"),
	}

	return settings, nil
}
