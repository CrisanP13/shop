package main

import (
	"context"
	"github.com/crisanp13/shop/src/api"
	"log"
)

func main() {
	logger := log.Default()
	port := ":8080"
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	if err := api.Run(logger, port, ctx); err != nil {
		log.Fatal("error on startup, ", err)
	}
}
