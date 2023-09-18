package configs

import "fmt"

// TODO export constants to a cfg.yml
const (
	APP_PORT = 8000

	DB_USER = "postgres"
	DB_PASS = "password"
	DB_NAME = "newsletter"
	DB_HOST = "localhost"
	DB_PORT = 5432
)

type DBSettings struct {
	user string
	pass string
	host string
	port uint16
	name string
}

func (db DBSettings) connection_string() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		db.user, db.pass, db.host, db.port, db.name,
	)
}

type AppSettings struct {
	database DBSettings
	port     uint16
}

func ConfigureApp() AppSettings {
	database := DBSettings{
		DB_USER,
		DB_PASS,
		DB_HOST,
		DB_PORT,
		DB_NAME,
	}

	settings := AppSettings{
		database,
		APP_PORT,
	}

	return settings
}
