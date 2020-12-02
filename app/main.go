package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver"
	metadatabox "github.com/sinmetalcraft/gcpbox/metadata"
	"github.com/sinmetalcraft/goma"
	v1 "github.com/sinmetalcraft/silverdile"
	v2 "github.com/sinmetalcraft/silverdile/v2"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	if metadatabox.OnGCP() {
		pID, err := metadatabox.ProjectID()
		if err != nil {
			panic(err)
		}
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: pID,
		})
		if err != nil {
			panic(err)
		}
		trace.RegisterExporter(exporter)
		trace.AlwaysSample()
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

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/image/resize/", v2hs.ResizeHandler)
	mux.HandleFunc("/v1", v1.ImageHandler)

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), &ochttp.Handler{
		Handler:     mux,
		Propagation: &b3.HTTPFormat{},
		FormatSpanName: func(req *http.Request) string {
			return fmt.Sprintf("/silverdile/%s", req.URL.Path)
		},
	}))
}
