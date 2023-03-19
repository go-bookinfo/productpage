package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
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

func main() {

	http.HandleFunc("/productpage", func(w http.ResponseWriter, r *http.Request) {
		var detail detail
		var review []review
		json.Unmarshal(getJson("http://detail/detail"), &detail)
		json.Unmarshal(getJson("http://review/review"), &review)

		t, _ := template.ParseFiles("index.html")
		t.Execute(w, Data{
			"detail": detail,
			"review": review,
		})

		// w.Write(detail)
		// w.Write(review)
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
