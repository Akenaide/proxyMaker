package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func getCardDeckInfo(url string) ([]Card, error) {
	var cardsDeck = []Card{}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return cardsDeck, err
	}
	doc.Find("div .wscard").Each(func(i int, s *goquery.Selection) {
		card := Card{}
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
