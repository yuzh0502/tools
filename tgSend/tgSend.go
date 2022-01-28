package tgSend

import (
	"net/http"
	"net/url"

	tgBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Send(proxyAddress string, chatID int64, text string) error {
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyAddress)
	}
	httpClient := &http.Client{Transport: &http.Transport{Proxy: proxy}}
	bot, err := tgBot.NewBotAPIWithClient("5009245437:AAG_AYeObETfG08YhLQNKGK-PAX1mowolY4", "https://api.telegram.org/bot%s/%s", httpClient)
	if err != nil {
		return err
	}
	msg := tgBot.NewMessage(chatID, text)
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
