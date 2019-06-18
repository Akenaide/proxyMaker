package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type wstcg struct{}

func (e wstcg) getCardDeckInfo(url string) ([]Card, error) {
	var cardsDeck = []Card{}
	return cardsDeck, nil
}

func (e wstcg) getCardDecksInfoList(url string) ([][]Card, error) {
	var decks = [][]Card{}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return decks, err
	}

	doc.Find("table.deckrecipe_table").Each(func(i int, s *goquery.Selection) {
		cardsDeck := []Card{}
		s.Find("tr.kind_2").Each(func(i int, s *goquery.Selection) {
			card := Card{}
			cardID := s.Find(".cardno").Text()
			if cardID != "" {
				var split = strings.Split(cardID, "-")
				card.ID = fmt.Sprintf("%s%s%s", split[0], "-", strings.Replace(split[1], "E", "", 1))
			} else {
				return
			}

			cardAmount := s.Find(".cardnum").Text()
			if cardAmount != "" {
				cardAmount = strings.TrimSpace(strings.TrimRight(cardAmount, "枚"))
				// Japanese bullshit '４' != '4'
				jpwhut := string(fmt.Sprintf("%+q", cardAmount))
				r := string(jpwhut[len(jpwhut)-2])
				card.Amount, err = strconv.Atoi(r)
				if err != nil {
					fmt.Println(err)
				}
			}
			cardsDeck = append(cardsDeck, card)
		})
		decks = append(decks, cardsDeck)
	})
	return decks, nil
}

func (e wstcg) isMine(url string) bool {
	return strings.Contains(url, "ws-tcg")
}
