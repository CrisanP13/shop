package main

import (
	"context"
	"log"
	"os"

	"github.com/crisanp13/shop/src/api"
)

func main() {
	logger := log.Default()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	if err := api.Run(ctx, os.Getenv, logger); err != nil {
		log.Fatal("error on startup, ", err)
	}
}
