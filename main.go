package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/vincent-petithory/dataurl"
)

const yuyuteiURL = "http://yuyu-tei.jp/"
const wsDeckUrl = "http://wsdecks.com/"

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
func lowCostSystemToURL(syspath string) string {
	return strings.Replace(syspath, "\\", "/", -1)
}

func convertToJpg(filePath string) {
	// convert -density 150 -trim to_love-ru_darkness_2nd_trial_deck.pdf -quality 100 -sharpen 0x1.0 love.jpg
	cmd := exec.Command("convert", "-density", "150", "-trim", filePath, "-quality", "100", "-sharpen", "0x1.0", filePath+".jpg")
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
	}
	err = cmd.Wait()
}

func main() {
	proxy := New("http://localhost:8080")
	os.MkdirAll(filepath.Join("static", "yuyutei"), 0744)
	os.MkdirAll(filepath.Join("static", "wsdeck"), 0744)
	// static := http.FileServer(http.Dir("./"))
	http.HandleFunc("/", proxy.handle)

	http.HandleFunc("/translationimages", func(w http.ResponseWriter, r *http.Request) {
		file := r.PostFormValue("file")
		filename := r.PostFormValue("filename")
		uid := strings.Replace(filename, filepath.Ext(filename), "", 1)
		dir := filepath.Join("static", uid)
		filePath := filepath.Join(dir, filename)

		data, err := dataurl.DecodeString(file)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// fmt.Println(dataURL.Data)
		os.MkdirAll(dir, 0777)
		ioutil.WriteFile(filePath, data.Data, 0644)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			convertToJpg(filePath)
		}
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		listJpg, err := filepath.Glob(filePath + "*.jpg")
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for i, jpgfile := range listJpg {
			listJpg[i] = lowCostSystemToURL(jpgfile)
		}

		b, err := json.Marshal(listJpg)

		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(b)
	})

	http.HandleFunc("/cardimages", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup

		result := []string{}
		link := r.PostFormValue("url")
		classCSS := r.PostFormValue("class_css")
		filter := ""

		if classCSS == "" {
			classCSS = ".card_list_box"
		}

		if link != "" {
			doc, err := goquery.NewDocument(link)
			uid := ""
			site := ""
			imageURL := ""

			if strings.Contains(link, yuyuteiURL) {
				site = "yuyutei"
				filter = classCSS + " .image img"
				parsedURL, _ := url.Parse(link)
				values, _ := url.ParseQuery(parsedURL.RawQuery)
				uid = values.Get("ver")
			} else if strings.Contains(link, wsDeckUrl) {
				site = "wsdeck"
				filter = ".wscard" + " img"
				uid = filepath.Base(link)
			}
			// currentDir, _ := os.Getwd()
			dir := filepath.Join("static", site, uid)

			if filter == "" {
				http.Error(w, fmt.Sprintln("Url is not supported", link), 500)
			}

			if err != nil {
				fmt.Println("Nope")
			}
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.MkdirAll(dir, 0744)
				doc.Find(filter).Each(func(i int, s *goquery.Selection) {
					wg.Add(1)
					val, _ := s.Attr("src")
					if site == "yuyutei" {
						big := strings.Replace(val, "90_126", "front", 1)
						imageURL = yuyuteiURL + big
					} else if site == "wsdeck" {
						imageURL = wsDeckUrl + val
					}

					go func(url string) {
						defer wg.Done()
						// fmt.Println("dir : ", dir)
						fileName := filepath.Join(dir, path.Base(url))
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
						// fmt.Println("image url: ", strings.Replace(fileName, "\\", "/", 1))
						result = append(result, lowCostSystemToURL(fileName))
					}(imageURL)
				})
			} else {
				files, err := ioutil.ReadDir(dir)
				if err != nil {
					fmt.Println(err)
				}
				for _, file := range files {
					absPath := filepath.Join(dir, file.Name())
					urlPath := lowCostSystemToURL(absPath)
					result = append(result, urlPath)
				}

			}
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

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("static", r.URL.Path[1:])
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":8010", nil)
}
