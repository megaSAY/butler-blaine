package main

import (
	"io/ioutil"
	"log"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/megasay/checkerr"
)

func startTelegramBot() {
	log.Printf("INFO:\tStarting Telegram bot.\n")

	token, err := ioutil.ReadFile(*flagTelegramToken)
	checkerr.Check(err, checkerr.ERROR, "Unable to read telegram bot token file.")
	bot, err := tgbotapi.NewBotAPI(string(token))
	checkerr.Check(err, checkerr.ERROR, "Unable create telegram bot, NewBotAPI error.")
	bot.Debug = false
	log.Printf("INFO:\tAuthorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if userAllowed(allowedIDs, update.Message.From.ID) {
			log.Println(update.Message.Text)
			answer := getResponse(update.Message.Text)
			log.Println(answer)
			//msg := getResponseMessage(update.Message.Text)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, answer))
		} else {
			log.Println("WARNING:\tDeined user. ID %i\n", update.Message.From.ID)
		}
	}
}

/*
func startTelegramBot(question chan<- string, answer chan<- string) {
	log.Printf("INFO:\tStarting Telegram bot.\n")

	token, err := ioutil.ReadFile(*flagTelegramToken)
	checkerr.Check(err, checkerr.ERROR, "Unable to read telegram bot token file.")
	bot, err := tgbotapi.NewBotAPI(string(token))
	checkerr.Check(err, checkerr.ERROR, "Unable create telegram bot, NewBotAPI error.")
	bot.Debug = false
	log.Printf("INFO:\tAuthorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if userAllowed(allowedIDs, update.Message.From.ID) {
			question <- update.Message.Text
			//msg := getResponseMessage(update.Message.Text)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "answer"))
		} else {
			log.Println("WARNING:\tDeined user. ID %i\n", update.Message.From.ID)
		}
	}
}
*/
