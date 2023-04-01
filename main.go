package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/opentracing/opentracing-go"
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
	http.HandleFunc("/productpage", func(w http.ResponseWriter, r *http.Request) {
		requestID, ctx, _ := Extract(r)
		fmt.Println(requestID, ctx)

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
