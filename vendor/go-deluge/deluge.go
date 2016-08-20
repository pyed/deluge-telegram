package deluge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
)

// Deluge represents an endpoint for Deluge RPC requests.
type Deluge struct {
	url      string
	password string

	client  *http.Client
	cookies []*http.Cookie

	id uint64
}

// New instantiates a new Deluge instance and authenticates with the
// server.
func New(url, password string) (*Deluge, error) {
	d := &Deluge{
		url,
		password,
		new(http.Client),
		nil,
		0,
	}

	d.client.Timeout = time.Duration(time.Second * 30)

	err := d.AuthLogin()
	if err != nil {
		return nil, err
	}

	return d, err
}

// GetTorrent takes a hash of a torrent to return *Torrent.
func (d *Deluge) GetTorrent(hash string) (*Torrent, error) {
	response, err := d.sendJsonRequest("core.get_torrent_status", []interface{}{hash, []string{}})
	if err != nil {
		return nil, err
	}

	torrent := new(Torrent)

	data, err := json.Marshal(response["result"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, torrent)
	if err != nil {
		return nil, err
	}

	if torrent.Hash == "" {
		return torrent, fmt.Errorf("No such torrent with hash: %s", hash)
	}

	return torrent, nil
}

// GetTorrents returns `Torrents` which is a slice of all available torrents.
func (d *Deluge) GetTorrents() (Torrents, error) {
	response, err := d.sendJsonRequest("core.get_torrents_status", []interface{}{nil, []string{}})
	if err != nil {
		return nil, err
	}

	jsonMap := response["result"].(map[string]interface{})
	torrents := make(Torrents, 0, len(jsonMap))

	for _, v := range jsonMap {
		torrent := new(Torrent)
		data, err := json.Marshal(v)
		if err != nil {
			return torrents, err
		}

		err = json.Unmarshal(data, torrent)
		if err != nil {
			return torrents, err
		}

		torrents = append(torrents, torrent)
	}

	// sort by name and add ids
	torrents.SortName(false)
	id := 1
	for _, torrent := range torrents {
		torrent.ID = id
		id++
	}

	return torrents, nil
}

// AddTorrentFile add torrent by file.
func (d *Deluge) AddTorrentFile(fileName, fileDump string, options map[string]interface{}) (string, error) {
	response, err := d.sendJsonRequest("core.add_torrent_file", []interface{}{fileName, fileDump, options})
	if err != nil {
		return "", err
	}

	if response["result"] == nil {
		return "", fmt.Errorf("Error adding: %s\nMaybe already added ?", fileName)
	}

	return response["result"].(string), nil
}

// AddTorrentMagnet adds a torrent via magnet url.
func (d *Deluge) AddTorrentMagnet(magnetUrl string) (string, error) {
	response, err := d.sendJsonRequest("core.add_torrent_magnet", []interface{}{magnetUrl, nil})
	if err != nil {
		return "", err
	}

	if response["result"] == nil {
		return "", fmt.Errorf("Error adding: %s\nMaybe already added ?", magnetUrl)
	}

	return response["result"].(string), nil
}

// AddTorrentUrl adds a torrent via http URL.
func (d *Deluge) AddTorrentUrl(torrentUrl string) (string, error) {
	response, err := d.sendJsonRequest("core.add_torrent_url", []interface{}{torrentUrl, nil})
	if err != nil {
		return "", err
	}
	if response["result"] == nil {
		return "", fmt.Errorf("Error adding: %s\nMaybe already added ?", torrentUrl)
	}
	return response["result"].(string), nil
}

// RemoveTorrent takes a hash of torrent to delete
func (d *Deluge) RemoveTorrent(hash string, removeData bool) error {
	// make sure that we have a torrent with the giving hash;
	// attempting to remove a hash that doesn't exists stalls for ever.
	if _, err := d.GetTorrent(hash); err != nil {
		return err
	}

	if _, err := d.sendJsonRequest("core.remove_torrent", []interface{}{hash, removeData}); err != nil {
		return err
	}

	return nil
}

// PauseTorrent takes a hash of a torrent to pause.
func (d *Deluge) PauseTorrent(hash string) error {
	if _, err := d.sendJsonRequest("core.pause_torrent", []interface{}{[]string{hash}}); err != nil {
		return err
	}

	return nil
}

// StartTorrent takes a hash of a torrent to start.
func (d *Deluge) StartTorrent(hash string) error {
	if _, err := d.sendJsonRequest("core.resume_torrent", []interface{}{[]string{hash}}); err != nil {
		return err
	}

	return nil
}

// PauseAll pauses all torrents.
func (d *Deluge) PauseAll() error {
	if _, err := d.sendJsonRequest("core.pause_all_torrents", []interface{}{}); err != nil {
		return err
	}
	return nil
}

// StartAll starts all torrents.
func (d *Deluge) StartAll() error {
	if _, err := d.sendJsonRequest("core.resume_all_torrents", []interface{}{}); err != nil {
		return err
	}
	return nil
}

// CheckTorrent takes a hash of a torrent to force re-check.
func (d *Deluge) CheckTorrent(hash string) error {
	if _, err := d.sendJsonRequest("core.force_recheck", []interface{}{[]string{hash}}); err != nil {
		return err
	}

	return nil
}

// SpeedRate returns download and upload speed in bytes.
func (d *Deluge) SpeedRate() (float64, float64, error) {
	response, err := d.sendJsonRequest("core.get_session_status",
		[]interface{}{[]string{"payload_download_rate", "payload_upload_rate"}})
	if err != nil {
		return -1, -1, err
	}

	data, err := json.Marshal(response["result"].(map[string]interface{}))
	if err != nil {
		return -1, -1, err
	}

	rate := &struct {
		Download float64 `json:"payload_download_rate"`
		Upload   float64 `json:"payload_upload_rate"`
	}{}

	err = json.Unmarshal(data, rate)
	if err != nil {
		return -1, -1, err
	}

	return rate.Download, rate.Upload, nil
}

// FilterTree wraps "get_filter_tree"
func (d *Deluge) FilterTree() ([][]interface{}, [][]interface{}, error) {
	response, err := d.sendJsonRequest("core.get_filter_tree", []interface{}{})
	if err != nil {
		return nil, nil, err
	}

	data, err := json.Marshal(response["result"].(map[string]interface{}))
	if err != nil {
		return nil, nil, err
	}

	tree := &struct {
		State       [][]interface{} `json:"state"`
		TrackerHost [][]interface{} `json:"tracker_host"`
	}{}

	err = json.Unmarshal(data, tree)
	if err != nil {
		return nil, nil, err
	}

	return tree.State, tree.TrackerHost, nil
}

// Version returns Deluge/libtorrent versions
func (d *Deluge) Version() (string, string, error) {
	response, err := d.sendJsonRequest("daemon.info", []interface{}{})
	if err != nil {
		return "", "", err
	}

	delugeVersion := response["result"].(string)

	response, err = d.sendJsonRequest("core.get_libtorrent_version", []interface{}{})
	if err != nil {
		return delugeVersion, "", err
	}

	libtorrentVersion := response["result"].(string)

	return delugeVersion, libtorrentVersion, nil
}

// AuthLogin gets called via New to authenticate with deluge.
func (d *Deluge) AuthLogin() error {
	response, err := d.sendJsonRequest("auth.login", []interface{}{d.password})
	if err != nil {
		return err
	}

	if response["result"] != true {
		return fmt.Errorf("authetication failed")
	}

	return nil
}

// sendJsonRequest takes a method and params to send to deluge and returns the output.
func (d *Deluge) sendJsonRequest(method string, params []interface{}) (map[string]interface{}, error) {
	atomic.AddUint64(&(d.id), 1)
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     d.id,
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", d.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if d.cookies != nil {
		for _, cookie := range d.cookies {
			req.AddCookie(cookie)
		}
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received non-ok status to http request : %d", resp.StatusCode)
	}

	d.cookies = resp.Cookies()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result["error"] != nil {
		return nil, fmt.Errorf("json error : %v", result["error"])
	}

	return result, err
}
