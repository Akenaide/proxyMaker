package main

type plugin interface {
	getCardDeckInfo(string) ([]Card, error)
	isMine(string) bool
}
