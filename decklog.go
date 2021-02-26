package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type decklog struct{}

type decklogCardInfo struct {
	CardID string `json:"card_number"`
	Amount int    `json:"num"`
}

type decklogResponse struct {
	Cards []decklogCardInfo `json:"list"`
}

func (e decklog) name() string {
	return "decklog"
}

func (e decklog) getCardDecksInfoList(url string) ([][]Card, error) {
	var decks = [][]Card{}
	info, err := e.getCardDeckInfo(url)
	if err != nil {
		fmt.Println(err)
	}
	decks = append(decks, info)
	return decks, nil
}

func (e decklog) getCardDeckInfo(url string) ([]Card, error) {
	var cardsDeck = []Card{}
	var decklogResp decklogResponse

	splits := strings.Split(url, ".com/")

	apiurl := fmt.Sprintf("%s.com/system/app/api/%s", splits[0], splits[1])
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiurl, nil)
	req.Header.Set("referer", url)

	resp, errresp := client.Do(req)

	if errresp != nil {
		fmt.Println("Error on resquest from decklog")
		return cardsDeck, errresp
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	parseerr := decoder.Decode(&decklogResp)

	if parseerr != nil {
		fmt.Println("Error on decode from decklog")
		return cardsDeck, parseerr
	}

	for _, info := range decklogResp.Cards {
		// card := Card{Amount: info.Amount, ID: info.CardID}
		cardsDeck = append(cardsDeck, Card{Amount: info.Amount, ID: info.CardID})
	}

	return cardsDeck, nil
}

func (e decklog) isMine(url string) bool {
	return strings.Contains(url, "decklog")
}
