package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
)

type dataTemplate struct {
	Name  string
	Cards []card
}

var cockatricCXMap = map[string]string{
	"CR": "R",
	"CU": "U",
	"CC": "C",
}

func getTranslationHotC(w http.ResponseWriter, r *http.Request) {
	fmt.Println("getTranslationHotC")
	var sem = make(chan bool, 2)

	translations := []Card{}
	cardsInfo, errGetCardDeckInfo := getCardDeckInfo(r.PostFormValue("url"))
	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	for _, card := range cardsInfo {
		sem <- true
		go func(card Card) {
			defer func() { <-sem }()
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
		}(card)
	}

	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	b, err := json.Marshal(translations)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(b)
}

func cardimages(w http.ResponseWriter, r *http.Request) {
	var link = r.PostFormValue("url")
	var cardsDeck = []Card{}
	var result = []string{}

	if link != "" {
	}
	cardsDeck, errGetCardDeckInfo := getCardDeckInfo(link)
	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	deck, err := getDeckConfig(link)
	if err != nil {
		fmt.Println(err)
	}

	createCardsCodeFile(deck.Dir, cardsDeck)
	for _, card := range cardsDeck {
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

func yytInfos(w http.ResponseWriter, r *http.Request) {
	fmt.Println("yytImage")
	out, err := os.Create(filepath.Join("static", "yyt_infos.json"))
	var buffer bytes.Buffer
	defer out.Close()
	cardMap := map[string]Card{}
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
				yytInfo := Card{URL: cardURL, Price: cardPrice}
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

func estimatePrice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("estimatePrice")
	var result = []Card{}
	var deckPrice int

	cardsInfo, errGetCardDeckInfo := getCardDeckInfo(r.PostFormValue("url"))
	if errGetCardDeckInfo != nil {
		fmt.Println(errGetCardDeckInfo)
	}

	for _, card := range cardsInfo {
		var total = card.Amount * yytMap[card.ID].Price
		deckPrice = deckPrice + total
		card.URL = yytMap[card.ID].URL
		card.Price = yytMap[card.ID].Price
		card.CardURL = yytMap[card.ID].CardURL
		result = append(result, card)
	}

	result = append(result, Card{ID: "TOTAL", Price: deckPrice})
	b, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(b)
}

func exportcockatrice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("exportcockatrice")
	var data = dataTemplate{}
	var b bytes.Buffer

	cardsInfo, errGetCardDeckInfo := getCardDeckInfo(r.PostFormValue("url"))
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
		io.WriteString(w, "ID is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	infos, exists := yytMap[ID[0]]
	if !exists {
		io.WriteString(w, fmt.Sprintf("%v does not exists", ID))
		w.WriteHeader(http.StatusBadRequest)
	} else {
		b, err := json.Marshal(infos)
		if err != nil {
			fmt.Println(err)
		}
		w.Write(b)
	}
}
