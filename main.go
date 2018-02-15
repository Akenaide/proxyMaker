package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const yuyuteiURL = "http://yuyu-tei.jp/"
const wsDeckURL = "https://wsdecks.com"
const hoTcURL = "http://www.heartofthecards.com/code/cardlist.html?card=WS_"
const yuyuteiBase = "http://yuyu-tei.jp/game_ws"

var yytMap = map[string]card{}

// Prox struct
type Prox struct {
	// target url of reverse proxy
	target *url.URL
	// instance of Go ReverseProxy thatwill do the job for us
	proxy *httputil.ReverseProxy
}

type site struct {
	Name   string
	Filter string
}

type deck struct {
	Dir  string
	Site site
}

type card struct {
	ID          string
	Translation string
	Amount      int
	URL         string
	Price       int
	CardURL     string
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

func createCardsCodeFile(dirPath string, cardsID []card) (string, error) {
	//TODO Do nothing if file exists
	os.MkdirAll(dirPath, 0744)
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

func getDeckConfig(link string) (deck, error) {
	uid := ""
	site := site{}

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
	deck := deck{Dir: dir, Site: site}
	if site.Filter == "" {
		return deck, fmt.Errorf("Url is not supported %v", link)
	}

	return deck, nil
}

func getCardDeckInfo(url string) ([]card, error) {
	fmt.Println("getCardDeckInfo")

	var cardsDeck = []card{}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return cardsDeck, err
	}
	doc.Find("div .wscard").Each(func(i int, s *goquery.Selection) {
		card := card{}
		cardID, exists := s.Attr("data-cardid")
		if exists {
			var split = strings.Split(cardID, "-")
			card.ID = fmt.Sprintf("%s%s%s", split[0], "-", strings.Replace(split[1], "E", "", 1))
		}

		cardAmount, exists := s.Attr("data-amount")
		if exists {
			card.Amount, err = strconv.Atoi(cardAmount)
			if err != nil {
				fmt.Println(err)
			}
		}
		cardsDeck = append(cardsDeck, card)
	})
	return cardsDeck, nil
}

func main() {
	proxy := New("http://localhost:8080")
	os.MkdirAll(filepath.Join("static", "yuyutei"), 0744)
	os.MkdirAll(filepath.Join("static", "wsdeck"), 0744)

	yytInfosData, yytErr := ioutil.ReadFile(filepath.Join("static", "yyt_infos.json"))
	if yytErr != nil {
		fmt.Println(yytErr)
	}
	json.Unmarshal(yytInfosData, &yytMap)

	http.HandleFunc("/", proxy.handle)

	http.HandleFunc("/views/translationimages", getTranslationHotC)
	http.HandleFunc("/views/cardimages", cardimages)
	http.HandleFunc("/views/estimateprice", estimatePrice)
	http.HandleFunc("/views/update_yyt_infos", yytInfos)
	http.HandleFunc("/views/searchcards", searchcards)

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("static", r.URL.Path[1:])
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.ListenAndServe(":8010", nil)
}
