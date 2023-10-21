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
	Database *DBSettings
	Redis    *RedisSettings
	Port     uint16
}

type DBSettings struct {
	user string
	pass string
	host string
	port uint16
	name string
}

func (db *DBSettings) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		db.user, db.pass, db.host, db.port, db.name,
	)
}

type RedisSettings struct {
	host string
	port string
	Conn string
}

func (r *RedisSettings) ConnectionString() string {
	return fmt.Sprintf("%s:%s", r.host, r.port)
}

func ConfigureApp() (settings *AppSettings, err error) {
	if e := viper.ReadInConfig(); e != nil {
		err = fmt.Errorf("failed to read configuration: %w", e)
		return
	}

	database := &DBSettings{
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetUint16("database.port"),
		viper.GetString("database.database_name"),
	}

	redis := &RedisSettings{
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.conn"),
	}

	port := viper.GetUint16("application_port")

	settings = &AppSettings{
		Database: database,
		Redis:    redis,
		Port:     port,
	}

	return
}

// EMAIL CLIENT
type EmailClientSettings struct {
	Server   string
	Port     int
	Username string
	Password string
	Sender   string
}

func ConfigureEmailClient(cfg string) (settings *EmailClientSettings, err error) {
	if cfg != "" {
		viper.SetConfigFile(cfg)
	}
	if e := viper.ReadInConfig(); e != nil {
		err = fmt.Errorf("failed to read configuration: %w", e)
		return
	}

	settings = &EmailClientSettings{
		viper.GetString("email.server"),
		viper.GetInt("email.port"),
		viper.GetString("email.username"),
		viper.GetString("email.password"),
		viper.GetString("email.sender"),
	}

	return
}
