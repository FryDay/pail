package pail

import (
	"github.com/FryDay/pail/sqlite"
	"github.com/bwmarrin/discordgo"
)

type Client struct {
	session *discordgo.Session
}

func NewClient(token, dbPath string) (*Client, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	db, err = sqlite.NewDB(dbPath)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages
	session.AddHandler(messageHandler)

	client := &Client{session: session}

	return client, nil
}

func (c *Client) Open() error {
	err := c.session.Open()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Close() {
	c.session.Close()
}
