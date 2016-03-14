package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type API struct {
	Message string `json:"message"`
}

const yuyuteiURL = "http://yuyu-tei.jp/"

func main() {
	http.HandleFunc("/cardimages", func(w http.ResponseWriter, r *http.Request) {
		link := r.PostFormValue("url")
		doc, err := goquery.NewDocument(link)
		parsedURL, _ := url.Parse(link)
		values, _ := url.ParseQuery(parsedURL.RawQuery)

		fmt.Println(values.Get("ver"))
		if err != nil {
			fmt.Println("Nope")
		}

		doc.Find(".card_list_box img").Each(func(i int, s *goquery.Selection) {
			val, _ := s.Attr("src")
			big := strings.Replace(val, "90_126", "front", 1)
			fmt.Printf("Link: n-%d __ %v%v\n", i, yuyuteiURL, big)
		})
	})
	http.ListenAndServe(":8010", nil)
}
