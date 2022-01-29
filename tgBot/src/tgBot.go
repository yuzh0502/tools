package src

import (
	"net/http"
	"net/url"

	tgBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	apiEndpoint = "https://api.telegram.org/bot%s/%s"
)

func getHttpClientWithProxy(proxyAddress string) *http.Client {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyAddress)
	}
	return &http.Client{Transport: &http.Transport{Proxy: proxy}}
}

func getNewBot() (*tgBot.BotAPI, error) {
	return tgBot.NewBotAPIWithClient(tgBotConfig.TgBotConfig.Token, apiEndpoint, getHttpClientWithProxy(tgBotConfig.TgBotConfig.Proxy))
}
