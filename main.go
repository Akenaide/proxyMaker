package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const yuyuteiURL = "http://yuyu-tei.jp/"

type Prox struct {
	// target url of reverse proxy
	target *url.URL
	// instance of Go ReverseProxy thatwill do the job for us
	proxy *httputil.ReverseProxy
}

func New(target string) *Prox {
	url, _ := url.Parse(target)
	// you should handle error on parsing
	return &Prox{target: url, proxy: httputil.NewSingleHostReverseProxy(url)}
}

func (p *Prox) handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-GoProxy", "GoProxy")
	// call to magic method from ReverseProxy object
	p.proxy.ServeHTTP(w, r)
}

func main() {
	// proxy := New("http://localhost:8080")
	static := http.FileServer(http.Dir("./"))

	http.Handle("/static", http.StripPrefix("/static", static))

	http.HandleFunc("/cardimages", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup

		result := []string{}
		link := r.PostFormValue("url")
		classCSS := r.PostFormValue("class_css")

		if classCSS == "" {
			classCSS = ".card_list_box"
		}

		if link != "" {
			doc, err := goquery.NewDocument(link)
			parsedURL, _ := url.Parse(link)
			values, _ := url.ParseQuery(parsedURL.RawQuery)
			uid := values.Get("ver")

			if err != nil {
				fmt.Println("Nope")
			}

			doc.Find(classCSS + " .image img").Each(func(i int, s *goquery.Selection) {
				wg.Add(1)
				val, _ := s.Attr("src")
				big := strings.Replace(val, "90_126", "front", 1)
				imageURL := yuyuteiURL + big

				go func(url string, dirName string) {
					defer wg.Done()
					currentDir, _ := os.Getwd()

					dir := filepath.Join(currentDir, dirName)
					fileName := filepath.Join(dirName, path.Base(url))
					os.MkdirAll(dir, 0744)
					out, err := os.Create(fileName)
					if err != nil {
						fmt.Println(err)
					}
					defer out.Close()
					reps, err := http.Get(url)

					if err != nil {
						fmt.Println(err)
					}

					file, err := io.Copy(out, reps.Body)
					if err != nil {
						fmt.Println(err)
					}

					fmt.Println("File", file)
					// fmt.Printf("Link: n-%d __ %v%v\n", i, imageURL, uid)
					defer reps.Body.Close()
					result = append(result, fileName)
				}(imageURL, uid)
			})
		}

		wg.Wait()
		fmt.Printf("Finish")
		b, err := json.Marshal(result)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(b)
	})
	// http.HandleFunc("/", proxy.handle)

	http.ListenAndServe(":8010", nil)
}
