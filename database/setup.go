package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

func InitializeAppDefault(ctx context.Context) *firebase.App {
	// opt := option.WithCredentialsFile(CredentialsPath)
	config := &firebase.Config{
		ProjectID:   ProjectID,
		DatabaseURL: DatabaseURL,
	}
	// [START initialize_app_default_golang]
	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	// [END initialize_app_default_golang]

	return app
}

func CreateDatabaseClient(ctx context.Context) *firestore.Client {
	app := InitializeAppDefault(ctx)
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln("Error initializing database client:", err)
	}
	return client
}
