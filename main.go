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

	humanize "github.com/dustin/go-humanize"

	deluge "go-deluge"

	"gopkg.in/telegram-bot-api.v4"
)

type View struct {
	Torrents deluge.Torrents
}

func (v *View) Update() (err error) {
	v.Torrents, err = Client.GetTorrents()
	return
}

func (v *View) GetTorrentByID(id int) (*deluge.Torrent, error) {
	// if there's no view, get one
	if view.Torrents == nil {
		if err := view.Update(); err != nil {
			log.Print("[ERROR] Deluge: %s", err)
			return nil, err
		}
	}

	for _, torrent := range v.Torrents {
		if torrent.ID == id {
			return torrent, nil
		}
	}
	return nil, fmt.Errorf("Can't find a torrent with ID: %d", id)
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

		case "downs", "/downs", "dl", "/dl":
			go downs(update)

		case "seeding", "/seeding", "sd", "/sd":
			go seeding(update)

		case "paused", "/paused", "pa", "/pa":
			go paused(update)

		case "checking", "/checking", "ch", "/ch":
			go checking(update)

		case "active", "/active", "ac", "/ac":
			go active(update)

		case "errors", "/errors", "er", "/er":
			go errors(update)

		// case "sort", "/sort", "so", "/so":
		// 	go sort(update, tokens[1:])

		case "trackers", "/trackers", "tr", "/tr":
			go trackers(update)

		case "add", "/add", "ad", "/ad":
			go add(update, tokens[1:])

		case "search", "/search", "se", "/se":
			go search(update, tokens[1:])

		case "latest", "/latest", "la", "/la":
			go latest(update, tokens[1:])

		case "info", "/info", "in", "/in":
			go info(update, tokens[1:])

		case "stop", "/stop", "sp", "/sp":
			go stop(update, tokens[1:])

		case "start", "/start", "st", "/st":
			go start(update, tokens[1:])

		case "check", "/check", "ck", "/ck":
			go check(update, tokens[1:])

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

		case "":
			// might be a file received
			go receiveTorrent(update)

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
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("list: %s"+err.Error(), ud.Message.Chat.ID, false)
		return
	}

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
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("list: %s"+err.Error(), ud.Message.Chat.ID, false)
		return
	}

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
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("list: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

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

// downs will send the names of torrents with status 'Downloading' or in queue to
func downs(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("list: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if torrent.State == "Downloading" {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		send("No downloads", ud.Message.Chat.ID, false)
		return
	}
	send(buf.String(), ud.Message.Chat.ID, false)
}

// seeding will send the names of the torrents with the status 'Seeding'
func seeding(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("seeding: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if torrent.State == "Seeding" {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		send("No torrents seeding", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// paused will send the names of the torrents with the status 'Seeding'
func paused(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("paused: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if torrent.State == "Paused" {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		send("No paused torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// checking will send the names of the torrents with the status 'Seeding'
func checking(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("checking: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if torrent.State == "Checking" {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		send("No torrents checking", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// active will send the names of the torrents with the status 'Seeding'
func active(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("active: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if torrent.DownloadPayloadRate > 0 ||
			torrent.UploadPayloadRate > 0 {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}

	if buf.Len() == 0 {
		send("No active torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// errors will send the names of the torrents with the status 'Seeding'
func errors(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("errors: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if !strings.Contains(torrent.TrackerStatus, "Announce OK") {
			buf.WriteString(fmt.Sprintf("<%d> %s\n%s\n", torrent.ID, torrent.Name, torrent.TrackerStatus))
		}
	}

	if buf.Len() == 0 {
		send("No errors torrents", ud.Message.Chat.ID, false)
		return
	}

	send(buf.String(), ud.Message.Chat.ID, false)

}

// trackers will send a list of trackers and how many torrents each one has
func trackers(ud tgbotapi.Update) {
	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("trackers: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	trackers := make(map[string]int)

	for _, torrent := range view.Torrents {
		if _, ok := trackers[torrent.TrackerHost]; !ok {
			trackers[torrent.TrackerHost] = 1
			continue
		}
		trackers[torrent.TrackerHost]++
	}

	buf := new(bytes.Buffer)
	for k, v := range trackers {
		buf.WriteString(fmt.Sprintf("%d - %s\n", v, k))
	}

	if buf.Len() == 0 {
		send("No trackers!", ud.Message.Chat.ID, false)
		return
	}
	send(buf.String(), ud.Message.Chat.ID, false)
}

// add takes an URL to a .torrent file to add it to transmission
func add(ud tgbotapi.Update, tokens []string) {
	if len(tokens) == 0 {
		send("add: needs atleast one URL", ud.Message.Chat.ID, false)
		return
	}

	var hash string
	var err error
	// loop over the URL/s and add them
	for _, url := range tokens {
		if strings.HasPrefix(url, "magnet") {
			hash, err = Client.AddTorrentMagnet(url)
		} else { // not a magnet
			hash, err = Client.AddTorrentUrl(url)
		}

		if err != nil {
			log.Printf("[ERROR] Deluge: %s", err)
			send(err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := Client.GetTorrent(hash)

		if err != nil {
			log.Printf("[ERROR] Deluge: %s", err)
			send("add: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		send(fmt.Sprintf("Added: %s", torrent.Name), ud.Message.Chat.ID, false)
	}
}

// receiveTorrent gets an update that potentially has a .torrent file to add
func receiveTorrent(ud tgbotapi.Update) {
	if ud.Message.Document.FileID == "" {
		return // has no document
	}

	// get the file ID and make the config
	fconfig := tgbotapi.FileConfig{
		FileID: ud.Message.Document.FileID,
	}
	file, err := Bot.GetFile(fconfig)
	if err != nil {
		send("receiver: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	// add by file URL
	add(ud, []string{file.Link(BotToken)})
}

// search takes a query and returns torrents with match
func search(ud tgbotapi.Update, tokens []string) {
	// make sure that we got a query
	if len(tokens) == 0 {
		send("search: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	query := strings.Join(tokens, " ")
	// "(?i)" for case insensitivity
	regx, err := regexp.Compile("(?i)" + query)
	if err != nil {
		send("search: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	if err := view.Update(); err != nil {
		log.Printf("[ERROR] Deluge: %s", err)
		send("search: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents {
		if regx.MatchString(torrent.Name) {
			buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
		}
	}
	if buf.Len() == 0 {
		send("No matches!", ud.Message.Chat.ID, false)
		return
	}
	send(buf.String(), ud.Message.Chat.ID, false)
}

// latest takes n and returns the latest n torrents
func latest(ud tgbotapi.Update, tokens []string) {
	var (
		n   = 5 // default to 5
		err error
	)

	if len(tokens) > 0 {
		n, err = strconv.Atoi(tokens[0])
		if err != nil {
			send("latest: argument must be a number", ud.Message.Chat.ID, false)
			return
		}
	}

	if err := view.Update(); err != nil {
		log.Print("[ERROR] Deluge: %s", err)
		send("latest: "+err.Error(), ud.Message.Chat.ID, false)
		return
	}
	// make sure that we stay in the boundaries
	if n <= 0 || n > len(view.Torrents) {
		n = len(view.Torrents)
	}

	// sort by age, and set reverse to true to get the latest first
	view.Torrents.SortAge(true)

	buf := new(bytes.Buffer)
	for _, torrent := range view.Torrents[:n] {
		buf.WriteString(fmt.Sprintf("<%d> %s\n", torrent.ID, torrent.Name))
	}
	if buf.Len() == 0 {
		send("latest: No torrents", ud.Message.Chat.ID, false)
		return
	}
	send(buf.String(), ud.Message.Chat.ID, false)
}

// info takes an id of a torrent and returns some info about it
func info(ud tgbotapi.Update, tokens []string) {
	if len(tokens) == 0 {
		send("info: needs a torrent ID number", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range tokens {
		torrentID, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("info: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := view.GetTorrentByID(torrentID)
		if err != nil {
			send(fmt.Sprintf("info: Can't find a torrent with an ID of %d", torrentID), ud.Message.Chat.ID, false)
			continue
		}

		// get an updated view of that torrent
		torrent, err = Client.GetTorrent(torrent.Hash)
		if err != nil {
			log.Printf("[ERROR] Deluge: %s", err)
			send("info: Deluge error while getting: "+torrent.Name, ud.Message.Chat.ID, false)
			continue
		}

		// format the info
		torrentName := mdReplacer.Replace(torrent.Name) // escape markdown
		info := fmt.Sprintf("`<%d>` *%s*\n%s (*%.1f%%*) ↓ *%s*  ↑ *%s* \nDL: *%s* UP: *%s* R: *%.3f*\nAdded: *%s*, ETA: *%d*\nTracker: `%s`",
			torrentID, torrentName, torrent.State, torrent.Progress,
			humanize.Bytes(uint64(torrent.DownloadPayloadRate)), humanize.Bytes(uint64(torrent.UploadPayloadRate)),
			humanize.Bytes(uint64(torrent.TotalPayloadDownload)), humanize.Bytes(uint64(torrent.TotalPayloadUpload)),
			torrent.Ratio, time.Unix(int64(torrent.TimeAdded), 0).Format(time.Stamp), torrent.Eta, torrent.TrackerHost)

		// send it
		msgID := send(info, ud.Message.Chat.ID, true)

		// this go-routine will make the info live for 'duration * interval'
		// takes torrent name so we don't have to use mdReplacer
		go func(torrentName string, torrentID, msgID int) {
			for i := 0; i < duration; i++ {
				time.Sleep(time.Second * interval)

				torrent, err = Client.GetTorrent(torrent.Hash)
				if err != nil {
					log.Printf("[ERROR] Deluge: %s", err)
					continue // skip this iteration if there's an error retrieving the torrent's info
				}

				info := fmt.Sprintf("`<%d>` *%s*\n%s (*%.1f%%*) ↓ *%s*  ↑ *%s* \nDL: *%s* UP: *%s* R: *%.3f*\nAdded: *%s*, ETA: *%d*\nTracker: `%s`",
					torrentID, torrentName, torrent.State, torrent.Progress,
					humanize.Bytes(uint64(torrent.DownloadPayloadRate)), humanize.Bytes(uint64(torrent.UploadPayloadRate)),
					humanize.Bytes(uint64(torrent.TotalPayloadDownload)), humanize.Bytes(uint64(torrent.TotalPayloadUpload)),
					torrent.Ratio, time.Unix(int64(torrent.TimeAdded), 0).Format(time.Stamp), torrent.Eta, torrent.TrackerHost)

				// update the message
				editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
				editConf.ParseMode = tgbotapi.ModeMarkdown
				Bot.Send(editConf)

			}

			// at the end write dashes to indicate that we are done being live.
			info := fmt.Sprintf("`<%d>` *%s*\n%s (*%.1f%%*) ↓ *-*  ↑ *-* \nDL: *%s* UP: *%s* R: *%.3f*\nAdded: *%s*, ETA: *-*\nTracker: `%s`",
				torrentID, torrentName, torrent.State, torrent.Progress, humanize.Bytes(uint64(torrent.TotalPayloadDownload)),
				humanize.Bytes(uint64(torrent.TotalPayloadUpload)), torrent.Ratio,
				time.Unix(int64(torrent.TimeAdded), 0).Format(time.Stamp), torrent.TrackerHost)

			editConf := tgbotapi.NewEditMessageText(ud.Message.Chat.ID, msgID, info)
			editConf.ParseMode = tgbotapi.ModeMarkdown
			Bot.Send(editConf)
		}(torrentName, torrentID, msgID)
	}
}

// stop takes id[s] of torrent[s] or 'all' to stop them
func stop(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("stop: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then stop all torrents
	if tokens[0] == "all" {
		if err := Client.PauseAll(); err != nil {
			send("stop: error occurred while stopping torrents", ud.Message.Chat.ID, false)
			return
		}
		send("stopped all torrents", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("stop: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := view.GetTorrentByID(num)
		if err != nil {
			send("stop: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		if err := Client.PauseTorrent(torrent.Hash); err != nil {
			log.Print("[ERROR] Deluge: %s", err)
			send("stop: an error occurred while stopping: "+torrent.Name, ud.Message.Chat.ID, false)
			continue
		}

		send(fmt.Sprintf("Stopped: %s", torrent.Name), ud.Message.Chat.ID, false)
	}
}

// start takes id[s] of torrent[s] or 'all' to start them
func start(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("start: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	// if the first argument is 'all' then start all torrents
	if tokens[0] == "all" {
		if err := Client.StartAll(); err != nil {
			log.Print("[ERROR] Deluge: %s", err)
			send("start: error occurred while starting some torrents", ud.Message.Chat.ID, false)
			return
		}
		send("started all torrents", ud.Message.Chat.ID, false)
		return

	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("start: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := view.GetTorrentByID(num)
		if err != nil {
			send("start: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		if err := Client.StartTorrent(torrent.Hash); err != nil {
			send("stop: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		send(fmt.Sprintf("Started: %s", torrent.Name), ud.Message.Chat.ID, false)
	}
}

// check takes id[s] of torrent[s] to verify them
func check(ud tgbotapi.Update, tokens []string) {
	// make sure that we got at least one argument
	if len(tokens) == 0 {
		send("check: needs an argument", ud.Message.Chat.ID, false)
		return
	}

	for _, id := range tokens {
		num, err := strconv.Atoi(id)
		if err != nil {
			send(fmt.Sprintf("check: %s is not a number", id), ud.Message.Chat.ID, false)
			continue
		}

		torrent, err := view.GetTorrentByID(num)
		if err != nil {
			send("check: "+err.Error(), ud.Message.Chat.ID, false)
			continue
		}

		if err := Client.CheckTorrent(torrent.Hash); err != nil {
			log.Print("[ERROR] Deluge: %s", err)
			send("check: ", ud.Message.Chat.ID, false)
			continue
		}

		send(fmt.Sprintf("Verifying: %s", torrent.Name), ud.Message.Chat.ID, false)
	}

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
