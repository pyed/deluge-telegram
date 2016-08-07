package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	deluge "github.com/pyed/go-deluge"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	// flags
	BotToken  string
	Master    string
	DelugeURL string
	Password  string
	LogFile   string

	// Deluge
	Client *deluge.Deluge

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
	log.Printf("[INFO] Token=%s\n\tMaster=%s\n\tURL=%s\n\tPASS=%s",
		BotToken, Master, DelugeURL, Password)
}

// init deluge
func init() {
	var err error
	Client, err = deluge.New(DelugeURL, Password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Deluge: %s", err)
		os.Exit(1)
	}
}

// init Telegram
func init() {
	var err error
	Bot, err = tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s", err)
		os.Exit(1)
	}
	log.Printf("[INFO] Authorized: %s", Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	Updates, err = Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Telegram: %s", err)
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
		// case "list", "/list", "li", "/li":
		// 	go list(update, tokens[1:])

		// case "head", "/head", "he", "/he":
		// 	go head(update, tokens[1:])

		// case "tail", "/tail", "ta", "/ta":
		// 	go tail(update, tokens[1:])

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
			// go send("no such command, try /help", update.Message.Chat.ID, false)

		}
	}
}
