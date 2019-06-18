package main

type plugin interface {
	getCardDecksInfoList(string) ([][]Card, error)
	getCardDeckInfo(string) ([]Card, error)
	isMine(string) bool
}
