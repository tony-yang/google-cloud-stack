package bookshelf

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"

	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	DB                BookDatabase
	OAuthConfig       *oauth2.Config
	SessionStore      sessions.Store
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
)

const (
	SQLPassword       string = "CHANGE ME"
	SQLInstance       string = "PROJECT:REGION:INSTANCE NAME"
	OAuthClientID     string = "OAuth ID"
	OAuthClientSecret string = "OAuth Secret"
	GCSBucketName     string = "Bucket Name"
	CookieSecret      string = "something-secret"
	oauthRedirectURL  string = "OAuth Redirect URL"
)

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func init() {
	var err error
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: "root",
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