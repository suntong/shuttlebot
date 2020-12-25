////////////////////////////////////////////////////////////////////////////
// Program: shuttlebot
// Purpose: Telegram Forwarding Shuttle Bot
// Authors: Tong Sun (c) 2018-2021, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/caarlos0/env"
	tb "gopkg.in/tucnak/telebot.v2"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

type config struct {
	LogLevel       string   `env:"SHUTTLEBOT_LOG"`
	TelegramToken  string   `env:"SHUTTLEBOT_TOKEN,required"`
	TelegramChatID []string `env:"SHUTTLEBOT_CID,required"`
}

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	progname = "shuttlebot"
	version  = "0.1.1"
	date     = "2020-12-25"

	logger log.Logger
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

//==========================================================================
// init

func init() {
	// https://godoc.org/github.com/go-kit/kit/log#TimestampFormat
	timestampFormat := log.TimestampFormat(time.Now, "0102T15:04:05") // 2006-01-02
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", timestampFormat)
}

//==========================================================================
// Main

func main() {
	app := Application{}
	app.Run()
}

////////////////////////////////////////////////////////////////////////////
// Application definition

// Application holds things together
type Application struct {
	Chat []*tb.Chat
	bot  *tb.Bot
}

// Run from start to end
func (app *Application) Run() {
	logger.Log("msg", "Telegram Forwarding Shuttle Bot",
		"version", version,
		"built-on", date,
	)

	c := config{}
	err := env.Parse(&c)
	abortOn("Config parsing error", err)

	bot, err := tb.NewBot(tb.Settings{
		Token:  c.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	abortOn("Can't start bot", err)
	app.bot = bot

	logger.Log("msg", "Copyright (C) 2018-2021, Tong Sun", "License", "MIT")
	app.Chat = make([]*tb.Chat, 0)
	for _, chat := range c.TelegramChatID {
		gi, err := strconv.ParseInt("-"+chat, 10, 64)
		abortOn("CID Parse error", err)
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

	logger.Log("msg", "Running with", "LogLevel", c.LogLevel)
	// if c.LogLevel == "Debug" {
	// 	log.SetLevel(log.DebugLevel)
	// }

	logger.Log("msg", "Bot started",
		"Bot", bot.Me.Username,
		"Forwarding-to", c.TelegramChatID[0],
	)
	bot.Start()
}

// ForwardHandler forwards received messages
func (app *Application) ForwardHandler(message *tb.Message) {
	// replyTo := fmt.Sprintf("%+v", message.ReplyTo)
	var replyTo string
	if message.ReplyTo == nil {
		replyTo = "null"
	} else {
		replyTo = message.ReplyTo.Text
	}
	logger.Log("msg", "Message received",
		"Sender", message.Sender.Recipient(),
		"Title", message.Chat.Title,
		"Text", message.Text,
		"ReplyTo", replyTo,
	)
	for _, chat := range app.Chat {
		app.bot.Forward(chat, message)
	}
}

//==========================================================================
// support functions

// abortOn will quit on anticipated errors gracefully without stack trace
func abortOn(errCase string, e error) {
	if e != nil {
		logger.Log("Abort", errCase, "Err", e)
	}
}
