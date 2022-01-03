package tgSend

import (
	tgBot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
	"net/http"
)

func Send(proxyAddress string, chatID int64, text string) error {
	dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	if err != nil {
		return err
	}
	httpClient := &http.Client{Transport: &http.Transport{Dial: dialer.Dial}}
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
