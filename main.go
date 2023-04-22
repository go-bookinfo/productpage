package main

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
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

func Init() (opentracing.Tracer, io.Closer) {
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	tracer, closer := jaeger.NewTracer(
		"",
		jaeger.NewConstSampler(false),
		jaeger.NewNullReporter(),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
	)
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

func Extract(r *http.Request) (string, opentracing.SpanContext, error) {
	requestID := r.Header.Get("x-request-id")
	spanCtx, err :=
		opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
	return requestID, spanCtx, err
}

func Inject(spanContext opentracing.SpanContext, request *http.Request, requestID string) error {
	request.Header.Add("x-request-id", requestID)
	return opentracing.GlobalTracer().Inject(
		spanContext,
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header))
}

func main() {
	_, closer := Init()
	defer closer.Close()
	http.HandleFunc("/productpage", func(w http.ResponseWriter, r *http.Request) {
		dc := make(chan []byte)
		rc := make(chan []byte)
		requestID, ctx, _ := Extract(r)

		var detail detail
		var review []review
		go getJson(dc, ctx, requestID, "http://detail/detail")
		go getJson(rc, ctx, requestID, "http://review/review")
		// json.Unmarshal(getJson(ctx, requestID, "http://detail/detail"), &detail)
		// json.Unmarshal(getJson(ctx, requestID, "http://review/review"), &review)
		json.Unmarshal(<-dc, &detail)
		json.Unmarshal(<-rc, &review)

		t, _ := template.ParseFiles("/app/index.html")
		t.Execute(w, Data{
			"detail": detail,
			"review": review,
		})
	})
	http.ListenAndServe(":80", nil)
}

func getJson(c chan []byte, ctx opentracing.SpanContext, requestID string, url string) {
	req, _ := http.NewRequest("GET", url, nil)
	Inject(ctx, req, requestID)

	// if err != nil {
	//     panic(err)
	// }

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// resp, err := http.Get(url)
	// if err != nil {
	// 	panic(err)
	// }
	// defer resp.Body.Close()

	json, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	c <- json
}
