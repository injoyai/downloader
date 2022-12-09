package bt

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Error(err)
	}
	T, err := c.AddTorrentFromFile("./test.torrent")
	if err != nil {
		t.Error(err)
	}
	T.DownloadAll()
}
