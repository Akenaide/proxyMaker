package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type encoredecks struct{}

type encoreCardInfo struct {
	Release string `json:"release"`
	Set     string `json:"set"`
	Sid     string `json:"sid"`
	Side    string `json:"side"`
}

type encoreResponse struct {
	Cards []encoreCardInfo `json:"cards"`
}

func encoreCompact(cards []encoreCardInfo) map[string]int {
	var res = map[string]int{}

	for _, _card := range cards {
		var name = fmt.Sprintf("%v/%v%v-%v", _card.Set, _card.Side, _card.Release, _card.Sid)
		res[name] = res[name] + 1
	}
	return res
}

func (e encoredecks) name() string {
	return "encoredeck"
}

func (e encoredecks) getCardDecksInfoList(url string) ([][]Card, error) {
	var decks = [][]Card{}
	info, err := e.getCardDeckInfo(url)
	if err != nil {
		fmt.Println(err)
	}
	decks = append(decks, info)
	return decks, nil
}

func (e encoredecks) getCardDeckInfo(url string) ([]Card, error) {
	var cardsDeck = []Card{}
	var encoreAPI = strings.Replace(url, "/deck", "/api/deck", 1)

	// TODO: Remove when cert will be good
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Get(encoreAPI)
	if err != nil {
		fmt.Println(err)
		return cardsDeck, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var encoreCards encoreResponse

	parseerr := decoder.Decode(&encoreCards)

	if parseerr != nil {
		fmt.Println("Error on decode from encoredeck")
		return cardsDeck, parseerr
	}

	for k, v := range encoreCompact(encoreCards.Cards) {
		card := Card{}
		card.ID = strings.ToUpper(k)
		card.Amount = v
		cardsDeck = append(cardsDeck, card)
	}

	return cardsDeck, nil
}

func (e encoredecks) isMine(url string) bool {
	return strings.Contains(url, "encoredecks")
}
