package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
)

const yuyuteiURL = "https://yuyu-tei.jp/"
const hoTcURL = "https://www.heartofthecards.com/code/cardlist.html?card=WS_"
const yuyuteiBase = "https://yuyu-tei.jp/game_ws"

var yytMap = map[string]Card{}
var plugins = []plugin{}

// New proxy
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
	proxy := New("http://localhost:8080")

	yytInfosData, yytErr := ioutil.ReadFile(filepath.Join("static", "yyt_infos.json"))
	if yytErr != nil {
		fmt.Println(yytErr)
	}
	json.Unmarshal(yytInfosData, &yytMap)

	plugins = append(plugins, encoredecks{})

	http.HandleFunc("/", proxy.handle)

	http.HandleFunc("/views/translationimages", getTranslationHotC)
	http.HandleFunc("/views/cardimages", cardimages)
	http.HandleFunc("/views/estimateprice", estimatePrice)
	http.HandleFunc("/views/searchcards", searchcards)
	http.HandleFunc("/views/exportcockatrice", exportcockatrice)

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":8010", nil)
}
