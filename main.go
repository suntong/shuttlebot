package main

import (
	"strconv"
	"time"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type config struct {
	LogLevel       string   `env:"SHUTTLEBOT_LOG"`
	TelegramToken  string   `env:"SHUTTLEBOT_TOKEN,required"`
	TelegramChatID []string `env:"SHUTTLEBOT_CID,required"`
}

var (
	progname = "shuttlebot"
	version  = "0.1.0"
	date     = "2018-12-31"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "0102 15:04:05", // 2006-01-02
		FullTimestamp:   true,
	})
}

func main() {
	app := Application{}
	app.Run()
}

// Application holds things together
type Application struct {
	Chat []*tb.Chat
	bot  *tb.Bot
}

// Run from start to end
func (app *Application) Run() {
	log.WithFields(log.Fields{
		"version":  version,
		"built-on": date,
	}).Info("Telegram Forwarding Shuttle Bot")

	c := config{}
	err := env.Parse(&c)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Config parsing error")
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:  c.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Can't start bot")
	}
	app.bot = bot

	log.WithFields(log.Fields{"License": "MIT"}).Info("Copyright (C) 2018-2019, Tong Sun")
	app.Chat = make([]*tb.Chat, 0)
	for _, chat := range c.TelegramChatID {
		gi, err := strconv.ParseInt("-"+chat, 10, 64)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("CID Parse error")
		}
		app.Chat = append(app.Chat, &tb.Chat{ID: gi})
	}

	bot.Handle(tb.OnText, app.ForwardHandler)
	bot.Handle(tb.OnAudio, app.ForwardHandler)
	bot.Handle(tb.OnContact, app.ForwardHandler)
	bot.Handle(tb.OnDocument, app.ForwardHandler)
	bot.Handle(tb.OnLocation, app.ForwardHandler)
	bot.Handle(tb.OnPhoto, app.ForwardHandler)
	bot.Handle(tb.OnVenue, app.ForwardHandler)
	bot.Handle(tb.OnVideo, app.ForwardHandler)
	bot.Handle(tb.OnVideoNote, app.ForwardHandler)
	bot.Handle(tb.OnVoice, app.ForwardHandler)
	// bot.Handle(tb.OnSticker, app.ForwardHandler)

	// bot.Handle(tb.OnPinned, savePinnedMessage)
	// bot.Handle(tb.OnAddedToGroup, showWelcomeMessage)

	log.WithFields(log.Fields{"LogLevel": c.LogLevel}).Info("Running with")
	if c.LogLevel == "Debug" {
		log.SetLevel(log.DebugLevel)
	} else if c.LogLevel == "Trace" {
		log.SetLevel(log.TraceLevel)
		log.SetReportCaller(true)
	}

	log.WithFields(log.Fields{
		"Bot":           bot.Me.Username,
		"Forwarding-to": c.TelegramChatID,
	}).Info("Bot started")
	bot.Start()
}

// ForwardHandler forwards received messages
func (app *Application) ForwardHandler(message *tb.Message) {
	log.WithFields(log.Fields{
		"Sender": message.Sender,
		"Title":  message.Chat.Title,
		"Text":   message.Text,
	}).Debug("Message received")
	for _, chat := range app.Chat {
		app.bot.Forward(chat, message)
	}
}
