package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/vincent-petithory/dataurl"
)

const yuyuteiURL = "http://yuyu-tei.jp/"
const wsDeckURL = "https://wsdecks.com"
const hoTcURL = "http://www.heartofthecards.com/code/cardlist.html?card=WS_"
const yuyuteiBase = "http://yuyu-tei.jp/game_ws"

// Prox struct
type Prox struct {
	// target url of reverse proxy
	target *url.URL
	// instance of Go ReverseProxy thatwill do the job for us
	proxy *httputil.ReverseProxy
}

type siteConfig struct {
	Name   string
	Filter string
}

type deckConfig struct {
	Dir  string
	Site siteConfig
}

type cardStruc struct {
	ID          string
	Translation string
}

type cardDeckInfo struct {
	ID     string
	Amount int
}

type YytInfo struct {
	URL   string
	Price int
}

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

func createCardsCodeFile(dirPath string, cardsID []cardDeckInfo) (string, error) {
	//TODO Do nothing if file exists
	dirPath += "/"
	out, err := os.Create(dirPath + "codes.txt")
	defer out.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	b, errMarshal := json.Marshal(cardsID)
	if errMarshal != nil {
		fmt.Println(errMarshal)
	}
	out.WriteString(string(b))
	return out.Name(), nil
}

func getTranslationHotC(codesPath string) []cardStruc {
	translations := []cardStruc{}
	file, err := os.Open(codesPath + "/codes.txt")
	scanner := bufio.NewScanner(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("getTranslationHotC")
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		cardsDeck := []cardDeckInfo{}
		json.Unmarshal(scanner.Bytes(), &cardsDeck)
		for _, cardDeck := range cardsDeck {
			url := hoTcURL + cardDeck.ID
			fmt.Println(url)
			doc, err := goquery.NewDocument(url)
			if err != nil {
				fmt.Println(err)
			}
			textHTML, err := doc.Find(".cards3").Slice(2, 3).Html()
			if err != nil {
				fmt.Println(err)
			}
			textHTML = strings.Replace(textHTML, "<br/>", "&#10;", -1)
			card := cardStruc{ID: cardDeck.ID, Translation: html.UnescapeString(textHTML)}

			translations = append(translations, card)
		}
	}
	return translations
}

func getDeckConfig(link string) (deckConfig, error) {
	uid := ""
	site := siteConfig{}

	if strings.Contains(link, yuyuteiURL) {
		site.Name = "yuyutei"
		site.Filter = ".card_list_box" + " .image img"
		parsedURL, _ := url.Parse(link)
		values, _ := url.ParseQuery(parsedURL.RawQuery)
		uid = values.Get("ver")
	} else if strings.Contains(link, wsDeckURL) {
		site.Name = "wsdeck"
		site.Filter = ".wscard" + " img"
		uid = filepath.Base(link)
	}
	dir := filepath.Join("static", site.Name, uid)
	deckConfig := deckConfig{Dir: dir, Site: site}
	if site.Filter == "" {
		return deckConfig, fmt.Errorf("Url is not supported %v", link)
	}

	return deckConfig, nil
}

func yytInfos(w http.ResponseWriter, r *http.Request) {
	fmt.Println("yytImage")
	out, err := os.Create(filepath.Join("static", "yyt_infos.json"))
	var buffer bytes.Buffer
	defer out.Close()
	cardMap := map[string]YytInfo{}
	filter := "ul[data-class=sell] .item_single_card .nav_list_second .nav_list_third a"
	doc, err := goquery.NewDocument(yuyuteiBase)

	if err != nil {
		fmt.Println("Error in get yyt urls")
	}

	doc.Find(filter).Each(func(i int, s *goquery.Selection) {
		url, has := s.Attr("href")
		fmt.Println(url)
		if has {
			images, errCard := goquery.NewDocument(yuyuteiURL + url)
			images.Find(".card_unit").Each(func(cardI int, cardS *goquery.Selection) {
				var price string
				price = cardS.Find(".price .sale").Text()
				if price == "" {
					price = strings.TrimSpace(cardS.Find(".price").Text())
				}
				cardPrice, errAtoi := strconv.Atoi(strings.TrimSuffix(price, "円"))
				if errAtoi != nil {
					fmt.Println(errAtoi)
				}
				cardURL, _ := cardS.Find(".image img").Attr("src")
				cardURL = strings.Replace(cardURL, "90_126", "front", 1)
				yytInfo := YytInfo{URL: cardURL, Price: cardPrice}
				cardMap[strings.TrimSpace(cardS.Find(".id").Text())] = yytInfo
			})
			if errCard != nil {
				fmt.Println(errCard)
			}
		}
	})
	b, errMarshal := json.Marshal(cardMap)
	if errMarshal != nil {
		fmt.Println(errMarshal)
	}
	json.Indent(&buffer, b, "", "\t")
	buffer.WriteTo(out)
	fmt.Println("finish")

}

func cardimages(w http.ResponseWriter, r *http.Request) {
	link := r.PostFormValue("url")
	cardIDs := []cardDeckInfo{}
	result := []string{}

	if link != "" {
	}
	doc, err := goquery.NewDocument(link)
	if err != nil {
		fmt.Println(err)
	}
	deckConfig, err := getDeckConfig(link)
	if err != nil {
		fmt.Println(err)
	}

	doc.Find("div .wscard").Each(func(i int, s *goquery.Selection) {
		card := cardDeckInfo{}
		cardID, exists := s.Attr("data-cardid")
		if exists {
			card.ID = cardID
		}

		cardAmount, exists := s.Attr("data-amount")
		if exists {
			card.Amount, err = strconv.Atoi(cardAmount)
			if err != nil {
				fmt.Println(err)
			}
		}
		cardIDs = append(cardIDs, card)
	})
	os.MkdirAll(deckConfig.Dir, 0744)
	createCardsCodeFile(deckConfig.Dir, cardIDs)
	yytInfos, yytErr := ioutil.ReadFile(filepath.Join("static", "yyt_infos.json"))
	if yytErr != nil {
		fmt.Println(yytErr)
	}
	var yytMap = map[string]string{}
	json.Unmarshal(yytInfos, &yytMap)
	for _, card := range cardIDs {
		cardURL, has := yytMap[card.ID]
		if has {
			urlPath := yuyuteiURL + cardURL
			result = append(result, urlPath)
		}
	}
	b, err := json.Marshal(result)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b)

}

func main() {
	proxy := New("http://localhost:8080")
	os.MkdirAll(filepath.Join("static", "yuyutei"), 0744)
	os.MkdirAll(filepath.Join("static", "wsdeck"), 0744)
	// static := http.FileServer(http.Dir("./"))
	http.HandleFunc("/", proxy.handle)

	http.HandleFunc("/translationimages", func(w http.ResponseWriter, r *http.Request) {
		link := r.PostFormValue("url")
		deckConfig, err := getDeckConfig(link)
		if err != nil {
			fmt.Println(err)
		}
		result := getTranslationHotC(deckConfig.Dir)
		b, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
		}
		w.Write(b)
	})

	http.HandleFunc("/translationimages_old", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/cardimages", cardimages)
	http.HandleFunc("/cardimages_old", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup

		result := []string{}
		link := r.PostFormValue("url")

		if link != "" {
			doc, err := goquery.NewDocument(link)
			imageURL := ""
			deckConfig, err := getDeckConfig(link)

			if err != nil {
				fmt.Println(err)
			}
			if _, err := os.Stat(deckConfig.Dir); os.IsNotExist(err) {
				os.MkdirAll(deckConfig.Dir, 0744)
				doc.Find(deckConfig.Site.Filter).Each(func(i int, s *goquery.Selection) {
					wg.Add(1)
					val, _ := s.Attr("src")
					if deckConfig.Site.Name == "yuyutei" {
						big := strings.Replace(val, "90_126", "front", 1)
						imageURL = yuyuteiURL + big
					} else if deckConfig.Site.Name == "wsdeck" {
						imageURL = wsDeckURL + val
					}

					go func(url string) {
						defer wg.Done()
						// fmt.Println("dir : ", dir)
						fileName := filepath.Join(deckConfig.Dir, path.Base(url))
						fileName = strings.Split(fileName, "?")[0]
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
				files, err := ioutil.ReadDir(deckConfig.Dir)
				if err != nil {
					fmt.Println(err)
				}
				yytInfos, yytErr := ioutil.ReadFile(filepath.Join("static", "yyt_infos.json"))
				if yytErr != nil {
					fmt.Println(yytErr)
				}
				var yytMap = map[string]string{}
				json.Unmarshal(yytInfos, &yytMap)
				for _, file := range files {
					cardID := strings.Replace(file.Name(), "_", "/", 1)
					cardID = strings.Replace(cardID, "_", "-", 1)
					cardID = strings.Split(cardID, ".")[0]
					fmt.Println(cardID)
					cardURL, has := yytMap[cardID]
					if has {
						urlPath := yuyuteiURL + cardURL
						result = append(result, urlPath)
					}
				}

			}
			wg.Wait()
			fmt.Printf("Finish")
			// createCardsCodeFile(deckConfig.Dir)
		}

		b, err := json.Marshal(result)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(b)
	})

	http.HandleFunc("/update_yyt_infos", yytInfos)

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("static", r.URL.Path[1:])
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":8010", nil)
}
