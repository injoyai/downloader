package bt

import "github.com/anacrolix/torrent"

type Client struct {
	*torrent.Client
}

func NewClient() (*Client, error) {
	c, err := torrent.NewClient(nil)
	return &Client{c}, err
}
