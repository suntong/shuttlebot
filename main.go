////////////////////////////////////////////////////////////////////////////
// Program: shuttlebot
// Purpose: Telegram Forwarding Shuttle Bot
// Authors: Tong Sun (c) 2018-2021, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	// 	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/caarlos0/env"
	tb "gopkg.in/tucnak/telebot.v2"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

const desc = "Telegram Forwarding Shuttle Bot"

type envConfig struct {
	TelegramToken string `env:"SHUTTLEBOT_TOKEN,required"`
	ConfigFile    string `env:"SHUTTLEBOT_CFG"`
	LogLevel      string `env:"SHUTTLEBOT_LOG"`

	TelegramChatID []string `env:"SHUTTLEBOT_CID,required"`
}

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	progname = "shuttlebot"
	version  = "2.0.0"
	date     = "2020-12-26"

	c      envConfig
	cfg    *Config
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
	err := env.Parse(&c)
	abortOn("Env config parsing error", err)
	if c.ConfigFile == "" {
		c.ConfigFile = "config.yaml"
	}
	cfg, err = getConfig(c.ConfigFile)
	abortOn("Config file reading error", err)
	//fmt.Printf("%#v\n", cfg)
	for ii := 0; ii < len(cfg.Forward); ii++ {
		sort.Strings(cfg.Forward[ii].To)
		sort.Ints(cfg.Forward[ii].User)
	}
	//fmt.Printf("%#v\n", cfg)

	fmt.Println(desc)
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
	logger.Log("msg", desc,
		"version", version,
		"built-on", date,
	)

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
	logger.Log("msg", "Message received",
		"Sender", message.Sender.Recipient(),
		"Title", message.Chat.Title,
		"Text", message.Text,
	)
	for _, chat := range app.Chat {
		if message.ReplyTo != nil {
			// replyTo, _ := json.Marshal(message.ReplyTo)
			// fmt.Println("ReplyTo:", string(replyTo))
			// if it replies to something, forward that first
			logger.Log("_replyto", message.ReplyTo.Text)
			app.bot.Forward(chat, message.ReplyTo)
		}
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
