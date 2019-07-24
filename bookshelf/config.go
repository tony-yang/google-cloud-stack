package bookshelf

import (
	"log"
	"os"
)

var (
	DB BookDatabase
)

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func init() {
	var err error
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: "root",
		Password: "12345",
		Instance: "ttyang-gcs:us-west1:library",
	})

	if err != nil {
		log.Fatal("cannot configure cloud SQL %v", err)
	}
}

func configureCloudSQL(c cloudSQLConfig) (BookDatabase, error) {
	if os.Getenv("GAE_INSTANCE") != "" {

		// Running in prod
		return newMySQLDB(MySQLConfig{
			Username:   c.Username,
			Password:   c.Password,
			UnixSocket: "/cloudsql/" + c.Instance,
		})
	}
	// Running locally
	return newMySQLDB(MySQLConfig{
		Username: c.Username,
		Password: c.Password,
		Host:     "localhost",
		Port:     3306,
	})
}