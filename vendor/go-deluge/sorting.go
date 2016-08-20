package deluge

import "sort"

type Sorting int

const (
	SortName Sorting = iota
	SortRevName
	SortAge
	SortRevAge
	SortSize
	SortRevSize
	SortProgress
	SortRevProgress
	SortDownSpeed
	SortRevDownSpeed
	SortUpSpeed
	SortRevUpSpeed
	SortDownloaded
	SortRevDownloaded
	SortUploaded
	SortRevUploaded
	SortRatio
	SortRevRatio
)

// sorting types
type (
	byName       Torrents
	byAge        Torrents
	bySize       Torrents
	byProgress   Torrents
	byDownSpeed  Torrents
	byUpSpeed    Torrents
	byDownloaded Torrents
	byUploaded   Torrents
	byRatio      Torrents
)

func (t byName) Len() int           { return len(t) }
func (t byName) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byName) Less(i, j int) bool { return t[i].Name < t[j].Name }

func (t byAge) Len() int           { return len(t) }
func (t byAge) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byAge) Less(i, j int) bool { return t[i].TimeAdded < t[j].TimeAdded }

func (t bySize) Len() int           { return len(t) }
func (t bySize) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t bySize) Less(i, j int) bool { return t[i].TotalSize < t[j].TotalSize }

func (t byProgress) Len() int           { return len(t) }
func (t byProgress) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byProgress) Less(i, j int) bool { return t[i].TotalDone < t[j].TotalDone }

func (t byDownSpeed) Len() int           { return len(t) }
func (t byDownSpeed) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byDownSpeed) Less(i, j int) bool { return t[i].DownloadPayloadRate < t[j].DownloadPayloadRate }

func (t byUpSpeed) Len() int           { return len(t) }
func (t byUpSpeed) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byUpSpeed) Less(i, j int) bool { return t[i].UploadPayloadRate < t[j].UploadPayloadRate }

func (t byDownloaded) Len() int           { return len(t) }
func (t byDownloaded) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byDownloaded) Less(i, j int) bool { return t[i].AllTimeDownload < t[j].AllTimeDownload }

func (t byUploaded) Len() int           { return len(t) }
func (t byUploaded) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byUploaded) Less(i, j int) bool { return t[i].TotalUploaded < t[j].TotalUploaded }

func (t byRatio) Len() int           { return len(t) }
func (t byRatio) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byRatio) Less(i, j int) bool { return t[i].Ratio < t[j].Ratio }

func (t Torrents) SortName(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byName(t)))
		return
	}
	sort.Sort(byName(t))
}

func (t Torrents) SortAge(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byAge(t)))
		return
	}
	sort.Sort(byAge(t))
}

func (t Torrents) SortSize(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(bySize(t)))
		return
	}
	sort.Sort(bySize(t))
}

func (t Torrents) SortProgress(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byProgress(t)))
		return
	}
	sort.Sort(byProgress(t))
}

func (t Torrents) SortDownSpeed(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byDownSpeed(t)))
		return
	}
	sort.Sort(byDownSpeed(t))
}

func (t Torrents) SortUpSpeed(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byUpSpeed(t)))
		return
	}
	sort.Sort(byUpSpeed(t))
}

func (t Torrents) SortDownloaded(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byDownloaded(t)))
		return
	}
	sort.Sort(byDownloaded(t))
}

func (t Torrents) SortUploaded(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byUploaded(t)))
		return
	}
	sort.Sort(byUploaded(t))
}

func (t Torrents) SortRatio(reverse bool) {
	if reverse {
		sort.Sort(sort.Reverse(byRatio(t)))
		return
	}
	sort.Sort(byRatio(t))
}
