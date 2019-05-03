package main

import (
	"net/http/httputil"
	"net/url"
)

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

type Card struct {
	ID          string
	Translation string
	Amount      int
	URL         string
	Price       int
	CardURL     string
	Rarity      string
}
