////////////////////////////////////////////////////////////////////////////
// Program: fetch.go
// Purpose: Telegram forwarding via fetching
// Authors: Tong Sun (c) 2021, All rights reserved
////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Fetch fetches the given media & send it back
func (app *Application) FetchHandler(message *tb.Message) {
	if !cfg.Fetchable {
		return
	}

	username := message.Sender.Username
	if len(username) == 0 {
		username = message.Sender.FirstName + " " + message.Sender.LastName
	}
	logIf(2, "Fetch-in-group",
		"group", message.Chat.Title, "user", username)
	if lacks(cfg.Fetch.Users, message.Sender.ID) {
		// Sender is not in the chosen User list
		logIf(2, "Ignored-sender", "group", "user", username)
		return
	}

	//url := strings.Replace(message.Text, cfg.Fetch.Command+" ", "", 1)
	logIf(3, "Fetch-cmd", "Text", message.Text, // "Url", url,
		"Payload", message.Payload)

	if debug >= 3 {
		app.bot.Send(message.Chat, "Rogher @"+username)
	}
	r := <-app.Fetch(message.Payload, message.Chat)
	if len(r) != 0 {
		app.bot.Send(message.Chat, r+" @"+username)
	}
}

// Fetch fetches from the given url and sent to TG group
func (app *Application) Fetch(url string, Chat *tb.Chat) chan string {
	r := make(chan string)
	//defer close(r)
	fileName := getFileName(url)
	if fileName == "" {
		r <- "Unknown url format"
		return r
	}

	fileName += ".webm"
	// download as fileName
	os.Chdir(cfg.Fetch.Folder)
	args := []string{"-o", fileName, url}
	args = append(args, cfg.Fetch.Vformat...)

	go func() {
		cmd := exec.Command(cfg.Fetch.Downloader, args...)
		err := cmd.Run()
		if err != nil {
			r <- err.Error()
			return
		}

		// get video duration
		cmdStr := "ffprobe -i " + fileName +
			" 2>&1 | sed -n '/^  *Duration: /{ s/^.*Duration: //; s/,.*$//; p; }'"
		out, err := exec.Command("sh", "-c", cmdStr).Output()
		if err != nil {
			r <- err.Error()
			return
		}
		d1, _ := time.Parse("15:04:05", string(out))
		d0, _ := time.Parse("15:04:05", "00:00:00")
		duration := int(d1.Sub(d0).Seconds())

		// Send out fileName
		logIf(2, "Send-video",
			"group", Chat.Title, "name", fileName)

		v := &tb.Video{
			File:  tb.FromDisk(fileName),
			Width: 640, Height: 360, Duration: duration,
		}
		if debug >= 3 {
			fmt.Printf("] %#v\n", v)
		}
		m, err := app.bot.Send(Chat, v)
		if err != nil {
			r <- err.Error()
			return
		}
		logIf(2, "Sent-video", "id", m.ID)

		r <- ""
	}()
	return r
}

func getFileName(url string) string {
	m := regexp.MustCompile(".*youtu.be/(.*)").FindStringSubmatch(url)
	if len(m) > 1 {
		return m[1]
	}
	m = regexp.MustCompile(`.*youtu.*\/watch\?v=(.*)`).FindStringSubmatch(url)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
