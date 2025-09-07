package bot

import (
	tele "gopkg.in/telebot.v4"
)

func NewBot(settings tele.Settings) (*tele.Bot, error) {
	b, err := tele.NewBot(settings)
	if err != nil {
		return nil, err
	}
	return b, err
}

func BindHandlers(bot *tele.Bot) {
	if bot == nil {
		return
	}


}
