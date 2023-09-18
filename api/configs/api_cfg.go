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
	Database DBSettings
	Port     uint16
}

func ConfigureApp() (DBSettings, uint16) {
	database := DBSettings{
		DB_USER,
		DB_PASS,
		DB_HOST,
		DB_PORT,
		DB_NAME,
	}

	port := uint16(APP_PORT)

	return database, port
}
