package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type dataTemplate struct {
	Name  string
	Cards []Card
}

type respMap struct {
	Resp *http.Response
	Card Card
}

var cockatricCXMap = map[string]string{
	"CR": "R",
	"CU": "U",
	"CC": "C",
}

const cacheTime = (time.Hour * 24) * 3

var findHotCImg = regexp.MustCompile(`/heart(.*).png`)

func getPlugin(url string) (plugin, error) {
	for _, plugin := range plugins {
		if plugin.isMine(url) {
			return plugin, nil
		}
	}

	return nil, fmt.Errorf("Url: (%v) not supported", url)
}

func fetchTranslation(cardsInfo []Card) []Card {
	translations := []Card{}

	for _, card := range cardsInfo {
		url := hoTcURL + card.ID + "&short=1"
		fmt.Println(url)
		doc, err := goquery.NewDocument(url)
		if err != nil {
			fmt.Println(err)
		}
		textHTML, err := doc.Find("body").Html()
		if err != nil {
			fmt.Println(err)
		}
		card.Translation = html.UnescapeString(textHTML)
		card.URL = yytMap[card.ID].URL

		translations = append(translations, card)
		time.Sleep(1 * time.Second)
	}

	return translations
}

func getTranslationHotC(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getTranslationHotC")
	var link = r.PostFormValue("url")
	plugin, err := getPlugin(link)
	var translations = []Card{}

	if err != nil {
		fmt.Println(err)
	}

	cardsInfo, errGetCardDeckInfo := plugin.getCardDeckInfo(link)
	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	code := strings.Split(link, "/")
	filename := filepath.Join("cache", plugin.name(), code[len(code)-1], "deck.json")

	if _, err := os.Stat(filename); err == nil {
		jsonFile, err := os.Open(filename)

		if err != nil {
			fmt.Println("file error:", err)
		}

		defer jsonFile.Close()

		bytesValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal(bytesValue, &translations)

	} else {
		translations = fetchTranslation(cardsInfo)

	}

	b, err := json.Marshal(translations)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(b)
}

func cardimages(w http.ResponseWriter, r *http.Request) {
	var link = r.PostFormValue("url")
	var result = []string{}
	plugin, err := getPlugin(link)

	if err != nil {
		fmt.Println(err)
	}

	cardsInfo, errGetCardDeckInfo := plugin.getCardDeckInfo(link)

	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	for _, card := range cardsInfo {
		card, has := yytMap[card.ID]
		if has {
			result = append(result, card.URL)
		}
	}
	b, err := json.Marshal(result)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(b)
}

func estimatePrice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("estimatePrice")
	var decks = [][]Card{}
	var link = r.PostFormValue("url")
	plugin, err := getPlugin(link)

	if err != nil {
		fmt.Println(err)
	}

	decksInfo, errGetCardDeckInfo := plugin.getCardDecksInfoList(link)
	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	for _, deck := range decksInfo {
		var result = []Card{}
		var deckPrice int
		for _, card := range deck {
			var total = card.Amount * yytMap[card.ID].Price
			deckPrice = deckPrice + total
			card.URL = yytMap[card.ID].URL
			card.Price = yytMap[card.ID].Price
			card.CardURL = yytMap[card.ID].CardURL
			result = append(result, card)
		}
		result = append(result, Card{ID: "TOTAL", Price: deckPrice})
		decks = append(decks, result)
	}

	b, err := json.Marshal(decks)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(b)
}

func exportcockatrice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("exportcockatrice")
	var data = dataTemplate{}
	var b bytes.Buffer
	var link = r.PostFormValue("url")
	plugin, err := getPlugin(link)

	if err != nil {
		fmt.Println(err)
	}

	cardsInfo, errGetCardDeckInfo := plugin.getCardDeckInfo(link)

	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	data.Name = "ex"
	for _, card := range cardsInfo {
		var complete = yytMap[card.ID]
		complete.Amount = card.Amount
		complete.ID = strings.Replace(complete.ID, "/", "-", 1)
		if elem, ok := cockatricCXMap[complete.Rarity]; ok {
			complete.Rarity = elem
		}
		data.Cards = append(data.Cards, complete)
	}

	t, err := template.ParseFiles("./cockatrice_template.xml")
	if err != nil {
		log.Println(err)
	}

	t.Execute(&b, data)

	if err != nil {
		fmt.Println(err)
	}
	w.Write(b.Bytes())
}
func searchcards(w http.ResponseWriter, r *http.Request) {
	fmt.Println("searchcards")
	ID, ok := r.URL.Query()["id"]

	if !ok {
		http.Error(w, "ID is empty", http.StatusBadRequest)
		return
	}

	infos, exists := yytMap[ID[0]]
	if !exists {
		http.Error(w, fmt.Sprintf("%v does not exists", ID), http.StatusBadRequest)
	} else {
		b, err := json.Marshal(infos)
		if err != nil {
			fmt.Println(err)
		}
		w.Write(b)
	}
}

func cache(w http.ResponseWriter, r *http.Request) {
	fmt.Println("cache")
	var links = r.PostFormValue("decks")
	var wg sync.WaitGroup

	for _, link := range strings.Split(links, ",") {

		plugin, err := getPlugin(links)
		if err != nil {
			fmt.Println(err)
		}

		code := strings.Split(link, "/")
		folderpath := filepath.Join("cache", plugin.name(), code[len(code)-1])
		filename := filepath.Join(folderpath, "deck.json")
		os.MkdirAll(folderpath, 744)

		if infoStat, err := os.Stat(filename); err == nil && r.URL.Query().Get("force") != "true" {
			isGoodEnough := time.Now().Sub(infoStat.ModTime()).Hours() < cacheTime.Hours()
			if isGoodEnough {
				fmt.Printf("USe cache %v for code: %v\n", infoStat.ModTime(), code)
				continue
			}
		}

		cardsInfo, errGetCardDeckInfo := plugin.getCardDeckInfo(link)
		if errGetCardDeckInfo != nil {
			fmt.Println(errGetCardDeckInfo)
		}

		var buffer bytes.Buffer

		translation := fetchTranslation(cardsInfo)

		for i, card := range translation {
			wg.Add(1)
			imgPath := filepath.Join(folderpath, filepath.Base(card.URL))
			newPath := strings.Join([]string{"https://proxyMaker.naide.moe", imgPath}, "/")

			fetchImgURL := string(card.URL)

			translation[i].Translation = findHotCImg.ReplaceAllString(card.Translation, newPath)

			go func(imgPath string, imgURL string) {
				for {

					out, err := os.Create(imgPath)
					if err != nil {
						time.Sleep(time.Second * 1)
						continue
					}

					defer out.Close()
					log.Println("get :", imgURL)
					resp, err := http.Get(imgURL)
					if err != nil {
						log.Println("retry get : ", imgURL)
						time.Sleep(time.Second * 1)
						continue
					}
					defer resp.Body.Close()
					io.Copy(out, resp.Body)
					if err != nil {
						time.Sleep(time.Second * 1)
						continue
					}

					wg.Done()
					break
				}
			}(imgPath, fetchImgURL)
			translation[i].URL = newPath
		}
		wg.Wait()
		marsh, err := json.Marshal(translation)

		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(500)
			break
		}

		out, err := os.Create(filename)
		if err != nil {
			log.Println("write error", err.Error())
			continue
		}
		json.Indent(&buffer, marsh, "", "\t")
		buffer.WriteTo(out)
		out.Close()

	}

	fmt.Println("end cache")

}
