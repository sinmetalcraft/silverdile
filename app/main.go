package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/sinmetalcraft/goma"
	v1 "github.com/sinmetalcraft/silverdile"
	v2 "github.com/sinmetalcraft/silverdile/v2"
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

	gomas := goma.NewStorageService(ctx, gcs)

	var v2hs *v2.ImageHandlers
	{
		is, err := v2.NewImageService(ctx, gcs, gomas)
		if err != nil {
			panic(err)
		}
		v2hs = v2.NewImageHandlers(ctx, "/v2/image", is)
	}

	log.Printf("Listening on port %s", port)
	http.HandleFunc("/v2/image/resize/", v2hs.ResizeHandler)
	http.HandleFunc("/v1", v1.ImageHandler)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), http.DefaultServeMux); err != nil {
		log.Printf("failed ListenAndServe err=%+v", err)
	}
}
