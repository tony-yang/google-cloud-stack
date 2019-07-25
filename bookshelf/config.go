package bookshelf

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

var (
	DB                BookDatabase
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func init() {
	var err error
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: "root",
		Password: "CHANGE ME",
		Instance: "ttyang-gcs:us-west1:library",
	})

	if err != nil {
		log.Fatalf("cannot configure cloud SQL %v", err)
	}

	StorageBucketName = "ttyang-gcs-library"
	StorageBucket, err = configureStorage(StorageBucketName)

	if err != nil {
		log.Fatalf("cannot configure storage bucket %v", err)
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

func configureStorage(bucketID string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucketID), nil
}