package deluge

type Torrents []*Torrent

type Torrent struct {
	ID int // custom
	// Comment             string  `json:"comment"`
	// ActiveTime          int     `json:"active_time"`
	// IsSeed              bool    `json:"is_seed"`
	Hash              string `json:"hash"`
	UploadPayloadRate int    `json:"upload_payload_rate"`
	// MoveCompletedPath   string  `json:"move_completed_path"`
	// Private             bool    `json:"private"`
	TotalPayloadUpload float64 `json:"total_payload_upload"`
	Paused             bool    `json:"paused"`
	// SeedRank            float64 `json:"seed_rank"`
	// SeedingTime         int     `json:"seeding_time"`
	// MaxUploadSlots      int     `json:"max_upload_slots"`
	// PrioritizeFirstLast bool    `json:"prioritize_first_last"`
	// DistributedCopies   float64 `json:"distributed_copies"`
	DownloadPayloadRate float64 `json:"download_payload_rate"`
	// Message             string  `json:"message"`
	// NumPeers            int     `json:"num_peers"`
	// MaxDownloadSpeed    int     `json:"max_download_speed"`
	// MaxConnections      int     `json:"max_connections"`
	// Compact             bool    `json:"compact"`
	Ratio float64 `json:"ratio"`
	// TotalPeers          int     `json:"total_peers"`
	TotalSize float64 `json:"total_size"`
	// TotalWanted         float64 `json:"total_wanted"`
	State string `json:"state"`
	// FilePriorities      []int   `json:"file_priorities"`
	// MaxUploadSpeed      int     `json:"max_upload_speed"`
	// RemoveAtRatio       bool    `json:"remove_at_ratio"`
	Tracker string `json:"tracker"`
	// SavePath            string  `json:"save_path"`
	Progress      float64 `json:"progress"`
	TimeAdded     float64 `json:"time_added"`
	TrackerHost   string  `json:"tracker_host"`
	TotalUploaded float64 `json:"total_uploaded"`
	// Files               []struct {
	// 	Index  int     `json:"index"`
	// 	Path   string  `json:"path"`
	// 	Offset float64 `json:"offset"`
	// 	Size   float64 `json:"size"`
	// } `json:"files"`
	TotalDone float64 `json:"total_done"`
	// NumPieces       int     `json:"num_pieces"`
	TrackerStatus string `json:"tracker_status"`
	// TotalSeeds      int     `json:"total_seeds"`
	// MoveOnCompleted bool    `json:"move_on_completed"`
	// NextAnnounce    int     `json:"next_announce"`
	// StopAtRatio     bool    `json:"stop_at_ratio"`
	// FileProgress        []float64     `json:"file_progress"`
	// MoveCompleted       bool          `json:"move_completed"`
	// PieceLength         float64       `json:"piece_length"`
	// AllTimeDownload     float64       `json:"all_time_download"`
	// MoveOnCompletedPath string        `json:"move_on_completed_path"`
	// NumSeeds            int           `json:"num_seeds"`
	// Peers               []interface{} `json:"peers"`
	Name string `json:"name"`
	// Trackers            []struct {
	// 	SendStats    bool   `json:"send_stats"`
	// 	Fails        int    `json:"fails"`
	// 	Verified     bool   `json:"verified"`
	// 	URL          string `json:"url"`
	// 	FailLimit    int    `json:"fail_limit"`
	// 	CompleteSent bool   `json:"complete_sent"`
	// 	Source       int    `json:"source"`
	// 	StartSent    bool   `json:"start_sent"`
	// 	Tier         int    `json:"tier"`
	// 	Updating     bool   `json:"updating"`
	// } `json:"trackers"`
	TotalPayloadDownload float64 `json:"total_payload_download"`
	// IsAutoManaged        bool    `json:"is_auto_managed"`
	// SeedsPeersRatio      float64 `json:"seeds_peers_ratio"`
	// Queue                int     `json:"queue"`
	// NumFiles             int     `json:"num_files"`
	Eta int `json:"eta"`
	// StopRatio            float64 `json:"stop_ratio"`
	// IsFinished           bool    `json:"is_finished"`
}
