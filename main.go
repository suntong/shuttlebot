////////////////////////////////////////////////////////////////////////////
// Program: shuttlebot
// Purpose: Telegram Forwarding Shuttle Bot
// Authors: Tong Sun (c) 2018-2021, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"github.com/caarlos0/env"
	tb "gopkg.in/tucnak/telebot.v2"
)

////////////////////////////////////////////////////////////////////////////
// Constant and data type/structure definitions

const desc = "Telegram Forwarding Shuttle Bot"

////////////////////////////////////////////////////////////////////////////
// Global variables definitions

var (
	progname = "shuttlebot"
	version  = "2.1.0"
	date     = "2021-03-28"
)

////////////////////////////////////////////////////////////////////////////
// Function definitions

//==========================================================================
// Main

func main() {
	// == Config handling
	err := env.Parse(&c)
	abortOn("Env config parsing error", err)
	if c.ConfigFile == "" {
		c.ConfigFile = "config.yaml"
	}
	if c.LogLevel != "" {
		di, err := strconv.ParseInt(c.LogLevel, 10, 8)
		abortOn("SHUTTLEBOT_LOG (int) parse error", err)
		debug = int(di)
	}
	cfg, err = getConfig(c.ConfigFile)
	abortOn("Config file reading error", err)
	//fmt.Printf("%#v\n", cfg)
	for ii := 0; ii < len(cfg.Forward); ii++ {
		cfg.FromGroups = append(cfg.FromGroups, cfg.Forward[ii].From)
		sort.Ints(cfg.Forward[ii].User)
		for _, chat := range cfg.Forward[ii].To {
			cfg.Forward[ii].Chat = append(cfg.Forward[ii].Chat, &tb.Chat{ID: -chat})
		}
	}
	sort.Ints(cfg.FromGroups)
	//fmt.Printf("%+v\n", cfg)

	// == Application start
	//fmt.Println(desc)
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

	logIf(0, "Copyright (C) 2018-2021, Tong Sun", "License", "MIT")
	bot.Handle(cfg.Fetch.Command, app.FetchHandler)
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

	logIf(0, "Running-with", "LogLevel", c.LogLevel)
	// if c.LogLevel == "Debug" {
	// 	log.SetLevel(log.DebugLevel)
	// }

	cfg.Fetchable = commandExists(cfg.Fetch.Downloader)
	cfg.Fetchable = cfg.Fetchable && commandExists(cfg.Fetch.Merger)
	cfg.Fetchable = cfg.Fetchable && commandExists("ffprobe")
	logIf(0, "Online-media", "Fetchable", cfg.Fetchable)

	logIf(0, "Bot-started.",
		"Bot", bot.Me.Username,
		"Watching", fmt.Sprintf("%v", cfg.FromGroups),
	)
	bot.Start()
}

// ForwardHandler forwards received messages
func (app *Application) ForwardHandler(message *tb.Message) {
	// https://pkg.go.dev/gopkg.in/tucnak/telebot.v2#Message
	if lacks(cfg.FromGroups, int(-message.Chat.ID)) {
		// message.Chat.ID is not from the watching groups, ignore
		logIf(3, "Ignored-from-group",
			"ID", message.Chat.ID,
			"name", message.Chat.Title)
		return
	}
	logMessageIf(3, message)
	//fmt.Printf("ReplyTo: %+v\n", message.ReplyTo)

	forwarded := false
	for _, fwd := range cfg.Forward {
		if lacks(fwd.User, message.Sender.ID) {
			// Sender is not in the chosen User list
			logIf(2, "Ignored-sender", "group", message.Chat.Title,
				"fname", message.Sender.FirstName, "lname", message.Sender.LastName)
			continue
		}
		if int(-message.Chat.ID) != fwd.From {
			logIf(3, "Skip none-matching group", "name", fwd.From)
			continue
		}
		for _, chat := range fwd.Chat {
			if message.ReplyTo != nil {
				// if it replies to something, forward that first
				logIf(1, "_Replyto", "Text", message.ReplyTo.Text)
				app.bot.Forward(chat, message.ReplyTo)
			}
			if l := len([]rune(message.Text)); l == 0 || l >= 2 {
				app.bot.Forward(chat, message)
			} else {
				logIf(2, "Ignored-short-message", "group", message.Chat.Title,
					"message", message.Text, "len", l)
			}
			forwarded = true
		}
	}
	if forwarded {
		logMessageIf(1, message)
	}
}

//==========================================================================
// support functions

// lacks searches for x in a sorted slice of ints and returns true
// if x is not present in a. The slice must be sorted in ascending order.
func lacks(a []int, x int) bool {
	ll := sort.SearchInts(a, int(x))
	return ll == len(a) || a[ll] != int(x)
}

// check if command exists
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	logIf(3, "Commandexist", "Cmd", cmd, "Exists", err == nil)
	return err == nil
}

func logMessageIf(level int, message *tb.Message) {
	// https://godoc.org/gopkg.in/tucnak/telebot.v2#Message
	logIf(level, "Received",
		"Group", message.Chat.Title,
		"Sender", message.Sender.Username,
		"Text", message.Text,
	)
}

// abortOn will quit on anticipated errors gracefully without stack trace
func abortOn(errCase string, e error) {
	if e != nil {
		logger.Log("Abort", errCase, "Err", e)
	}
}
