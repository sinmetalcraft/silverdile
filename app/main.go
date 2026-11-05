package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/sinmetalcraft/silverdile/v3"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	v3ImageService, err := silverdile.NewImageService(ctx, gcs)
	if err != nil {
		panic(err)
	}
	v3hs := silverdile.NewResizeHandlers(ctx, "/v3/image", "alter-sinmetal-ci-silverdile", v3ImageService)

	mux := http.NewServeMux()
	mux.HandleFunc("/v3/image/resize/", v3hs.ResizeHandler)

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}
