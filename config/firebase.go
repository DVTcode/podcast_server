package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitFirebase() *firebase.App {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile("credentials/google-credentials.json"))
	if err != nil {
		log.Fatalf("error initializing Firebase: %v", err)
	}
	return app
}
