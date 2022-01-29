package main

import (
	"net/http"
	"net/url"

	tgBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	token       = "5009245437:AAG_AYeObETfG08YhLQNKGK-PAX1mowolY4"
	apiEndpoint = "https://api.telegram.org/bot%s/%s"
	adminID     = "956772010"
)

func getHttpClientWithProxy(proxyAddress string) *http.Client {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyAddress)
	}
	return &http.Client{Transport: &http.Transport{Proxy: proxy}}
}

func getNewBot() (*tgBot.BotAPI, error) {
	return tgBot.NewBotAPIWithClient(token, apiEndpoint, getHttpClientWithProxy(proxyAddress))
}
