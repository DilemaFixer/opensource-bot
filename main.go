package main

import (
	"log"
	"time"

	tele "gopkg.in/telebot.v4"
	"opensource-bot/bot"
)

func main() {
	pref := tele.Settings{
		Token:  "Token",
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := bot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.BindHandlers(b)
	b.Start()
}
