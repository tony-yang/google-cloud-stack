package bookshelf

import (
	"context"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"

	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	DB                BookDatabase
	OAuthConfig       *oauth2.Config
	PubsubClient      *pubsub.Client
	SessionStore      sessions.Store
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

var (
	ProjectID         string = "rw-bookshelf"
	SQLUser           string = strings.TrimSuffix(os.Getenv("DB_USER"), "\n")
	SQLPassword       string = strings.TrimSuffix(os.Getenv("DB_PASSWORD"), "\n")
	SQLInstance       string = "rw-bookshelf:us-west1:library"
	OAuthClientID     string = strings.TrimSuffix(os.Getenv("OAUTH"), "\n")
	OAuthClientSecret string = strings.TrimSuffix(os.Getenv("SECRET"), "\n")
	GCSBucketName     string = "rw-bookshelf-library"
	CookieSecret      string = "something-secret"
	oauthRedirectURL  string = "http://" + strings.TrimSuffix(os.Getenv("REDIRECT"), "\n") + "/oauth2callback"
	PubsubTopicID     string = "fill-book-details"
)

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func init() {
	var err error
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: SQLUser,
		Password: SQLPassword,
		Instance: SQLInstance,
	})

	if err != nil {
		log.Fatalf("cannot configure cloud SQL %v", err)
	}

	StorageBucketName = GCSBucketName
	StorageBucket, err = configureStorage(StorageBucketName)

	if err != nil {
		log.Fatalf("cannot configure storage bucket %v", err)
	}

	OAuthConfig = configureOAuthClient(OAuthClientID, OAuthClientSecret)

	cookieStore := sessions.NewCookieStore([]byte(CookieSecret))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore

	PubsubClient, err = configurePubsub(ProjectID)
	if err != nil {
		log.Fatal(err)
	}
}

func configureCloudSQL(c cloudSQLConfig) (BookDatabase, error) {
	if os.Getenv("GAE_INSTANCE") != "" {
		// Running in GAE
		return newMySQLDB(MySQLConfig{
			Username:   c.Username,
			Password:   c.Password,
			UnixSocket: "/cloudsql/" + c.Instance,
		})
	}
	// Running through the cloud_sql_proxy
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

func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = oauthRedirectURL
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func configurePubsub(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create the topic if it doesn't exist.
	if exists, err := client.Topic(PubsubTopicID).Exists(ctx); err != nil {
		return nil, err
	} else if !exists {
		if _, err := client.CreateTopic(ctx, PubsubTopicID); err != nil {
			return nil, err
		}
	}
	return client, nil
}
