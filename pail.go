package pail

import (
	"github.com/FryDay/pail/sqlite"
	"github.com/bwmarrin/discordgo"
)

type Pail struct {
	session *discordgo.Session
	db      *sqlite.DB
}

func NewPail(token, dbPath string) (*Pail, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return nil, err
	}
	client := &Pail{
		session: session,
		db:      db,
	}

	client.session.Identify.Intents = discordgo.IntentsGuildMessages
	client.session.AddHandler(client.messageHandler)

	return client, nil
}

func (p *Pail) Open() error {
	return p.session.Open()
}

func (p *Pail) Close() {
	p.session.Close()
}
