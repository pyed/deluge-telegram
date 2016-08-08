package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	deluge "go-deluge"

	"gopkg.in/telegram-bot-api.v4"
)

type View struct {
	Torrents []*deluge.Torrent
}

func (v *View) Update() (err error) {
	v.Torrents, err = Client.GetTorrents()
	return
}

var (
	// flags
	BotToken  string
	Master    string
	DelugeURL string
	Password  string
	LogFile   string

	// Deluge
	Client *deluge.Deluge

	// Deluge view
	view = new(View)

	// Telegram
	Bot     *tgbotapi.BotAPI
	Updates <-chan tgbotapi.Update

	// interval in seconds for live updates, affects: "active", "info", "speed", "head", "tail"
	interval time.Duration = 2
	// duration controls how many intervals will happen
	duration = 60

	// since telegram's markdown can't be escaped, we have to replace some chars
	mdReplacer = strings.NewReplacer("*", "•",
		"[", "(",
		"]", ")",
		"_", "-",
		"`", "'")
)

// init flags
func init() {
	// define arguments and parse them.
	flag.StringVar(&BotToken, "token", "", "Telegram bot token")
	flag.StringVar(&Master, "master", "", "Your telegram handler, So the bot will only respond to you")
	flag.StringVar(&DelugeURL, "url", "http://localhost:8112", "Deluge WebUI URL")
	flag.StringVar(&Password, "password", "deluge", "Deluge WebUI password")
	flag.StringVar(&LogFile, "logfile", "", "Send logs to a file")

	// set the usage message
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: deluge-telegram -token=<TOKEN> -master=<@tuser> -url=[http://] -password=[pass]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// make sure that we have the two madatory arguments: telegram token & master's handler.
	if BotToken == "" ||
		Master == "" {
		fmt.Fprintf(os.Stderr, "Error: Mandatory argument missing! (-token or -master)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// make sure that the handler doesn't contain @
	Master = strings.Replace(Master, "@", "", -1)

	// if we got a log file, log to it
	if LogFile != "" {
		logf, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logf)
	}
	// log the flags
	log.Printf(`[INFO] Token  = %s
			   Master = %s
			   URL    = %s
		   	   PASS   = %s`,
		BotToken, Master, DelugeURL, Password)
}

// init deluge
func init() {
	var err error
	Client, err = deluge.New(DelugeURL+"/json", Password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Deluge: %s\n", err)
		os.Exit(1)
	}

	// get a view
	if err := view.Update(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Deluge: %s\n", err)
		os.Exit(1)
	}

}

// init Telegram
func init() {
	var err error
	Bot, err = tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s\n", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Authorized: %s", Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	Updates, err = Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	for update := range Updates {
		// ignore edited messages
		if update.Message == nil {
			continue
		}

		// ignore anyone other than 'master'
		if strings.ToLower(update.Message.From.UserName) != strings.ToLower(Master) {
			log.Printf("[INFO] Ignored a message from: %s", update.Message.From.String())
			continue
		}

		// tokenize the update
		tokens := strings.Split(update.Message.Text, " ")
		command := strings.ToLower(tokens[0])

		switch command {
		case "update", "/update", "ud", "/ud":
			view.Update()

		case "list", "/list", "li", "/li":
			go list(update, tokens[1:])

		case "head", "/head", "he", "/he":
			go head(update, tokens[1:])

		case "tail", "/tail", "ta", "/ta":
			go tail(update, tokens[1:])

		// case "downs", "/downs", "dl", "/dl":
		// 	go downs(update)

		// case "seeding", "/seeding", "sd", "/sd":
		// 	go seeding(update)

		// case "paused", "/paused", "pa", "/pa":
		// 	go paused(update)

		// case "checking", "/checking", "ch", "/ch":
		// 	go checking(update)

		// case "active", "/active", "ac", "/ac":
		// 	go active(update)

		// case "errors", "/errors", "er", "/er":
		// 	go errors(update)

		// case "sort", "/sort", "so", "/so":
		// 	go sort(update, tokens[1:])

		// case "trackers", "/trackers", "tr", "/tr":
		// 	go trackers(update)

		// case "add", "/add", "ad", "/ad":
		// 	go add(update, tokens[1:])

		// case "search", "/search", "se", "/se":
		// 	go search(update, tokens[1:])

		// case "latest", "/latest", "la", "/la":
		// 	go latest(update, tokens[1:])

		// case "info", "/info", "in", "/in":
		// 	go info(update, tokens[1:])

		// case "stop", "/stop", "sp", "/sp":
		// 	go stop(update, tokens[1:])

		// case "start", "/start", "st", "/st":
		// 	go start(update, tokens[1:])

		// case "check", "/check", "ck", "/ck":
		// 	go check(update, tokens[1:])

		// case "stats", "/stats", "sa", "/sa":
		// 	go stats(update)

		// case "speed", "/speed", "ss", "/ss":
		// 	go speed(update)

		// case "count", "/count", "co", "/co":
		// 	go count(update)

		// case "del", "/del":
		// 	go del(update, tokens[1:])

		// case "deldata", "/deldata":
		// 	go deldata(update, tokens[1:])

		// case "help", "/help":
		// 	go send(HELP, update.Message.Chat.ID, true)

		// case "version", "/version":
		// 	go version(update)

		// case "":
		// 	// might be a file received
		// 	go receiveTorrent(update)

		default:
			// no such command, try help
			go send("no such command, try /help", update.Message.Chat.ID, false)

		}
	}
}

// list will form and send a list of all the torrents
// takes an optional argument which is a query to match against trackers
// to list only torrents that has a tracker that matchs.
func list(ud tgbotapi.Update, tokens []string) {
	buf := new(bytes.Buffer)
	// if it gets a query, it will list torrents that has trackers that match the query
	if len(tokens) != 0 {
		// (?i) for case insensitivity
		regx, err := regexp.Compile("(?i)" + tokens[0])
		if err != nil {
			log.Printf("[ERROR] Regexp: %s", err)
			send("list: "+err.Error(), ud.Message.Chat.ID, false)
			return
		}

		for _, torrent := range view.Torrents {
			if regx.MatchString(torrent.TrackerHost) {
				buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
			}
		}
	} else { // if we did not get a query, list all torrents
		for _, torrent := range view.Torrents {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		if len(tokens) != 0 { // if we got a tracker query show different message
			send(fmt.Sprintf("list: No tracker matches: *%s*", tokens[0]),
				ud.Message.Chat.ID, true)
			return
		}
		send("list: No torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)
}

// head will list the first 5 or n torrents
func head(ud tgbotapi.Update, tokens []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(tokens) > 0 {
		n, err = strconv.Atoi(tokens[0])
		if err != nil {
			send("head: argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	// make sure that we stay in the boundaries
	if n <= 0 || n > len(view.Torrents) {
		n = len(view.Torrents)
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents[:n] {
		buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
	}

	if buf.Len() == 0 {
		send("head: No torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// tail will list the first 5 or n torrents
func tail(ud tgbotapi.Update, tokens []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(tokens) > 0 {
		n, err = strconv.Atoi(tokens[0])
		if err != nil {
			send("tail: argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	// make sure that we stay in the boundaries
	if n <= 0 || n > len(view.Torrents) {
		n = len(view.Torrents)
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents[len(view.Torrents)-n:] {
		buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
	}

	if buf.Len() == 0 {
		send("tail: No torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// send takes a chat id and a message to send, returns the message id of the send message
func send(text string, chatID int64, markdown bool) int {
	// set typing action
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	Bot.Send(action)

	// check the rune count, telegram is limited to 4096 chars per message;
	// so if our message is > 4096, split it in chunks the send them.
	msgRuneCount := utf8.RuneCountInString(text)
LenCheck:
	stop := 4095
	if msgRuneCount > 4096 {
		for text[stop] != 10 { // '\n'
			stop--
		}
		msg := tgbotapi.NewMessage(chatID, text[:stop])
		msg.DisableWebPagePreview = true
		if markdown {
			msg.ParseMode = tgbotapi.ModeMarkdown
		}

		// send current chunk
		if _, err := Bot.Send(msg); err != nil {
			log.Printf("[ERROR] Send: %s", err)
		}
		// move to the next chunk
		text = text[stop:]
		msgRuneCount = utf8.RuneCountInString(text)
		goto LenCheck
	}

	// if msgRuneCount < 4096, send it normally
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	if markdown {
		msg.ParseMode = tgbotapi.ModeMarkdown
	}

	resp, err := Bot.Send(msg)
	if err != nil {
		log.Printf("[ERROR] Send: %s", err)
	}

	return resp.MessageID
}
