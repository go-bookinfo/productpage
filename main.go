package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Data map[string]interface{}

type detail struct {
	Name      string
	Summary   string
	Type      string
	Page      int
	Publisher string
	Language  string
	Isbn10    string
	Isbn13    string
}

type review struct {
	Id       int
	Star     int
	Reviewer string
	Review   string
	Color    string
}

const (
	service     = "trace-demo"
	environment = "production"
	id          = 1
)

func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
	return tp, nil
}

func main() {

	// for jeager
	tp, err := tracerProvider("http://jaeger-collector.istio-system:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("component-main")

	http.HandleFunc("/productpage", func(w http.ResponseWriter, r *http.Request) {
		var detail detail
		var review []review
		json.Unmarshal(getJson("http://detail/detail"), &detail)
		json.Unmarshal(getJson("http://review/review"), &review)
		fmt.Println(detail)
		fmt.Println(review)

		t, _ := template.ParseFiles("/app/index.html")
		t.Execute(w, Data{
			"detail": detail,
			"review": review,
		})

		// w.Write(detail)
		// w.Write(review)
		_, span := tr.Start(ctx, "productpage")
		defer span.End()
	})
	http.ListenAndServe(":80", nil)
}

func getJson(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return json
}
