package main

import ()

type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}
