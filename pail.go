package pail

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/FryDay/pail/sqlite"
	"github.com/bwmarrin/discordgo"
)

type Pail struct {
	config       *Config
	session      *discordgo.Session
	db           *sqlite.DB
	lastFact     *Fact
	randomTicker *time.Ticker
	randomReset  chan bool
}

func NewPail(config *Config, dbPath string) (*Pail, error) {
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, err
	}
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		return nil, err
	}
	client := &Pail{
		config:  config,
		session: session,
		db:      db,
	}

	client.session.Identify.Intents = discordgo.IntentsGuildMessages
	client.session.AddHandler(client.messageHandler)

	client.randomTicker = time.NewTicker(time.Minute * time.Duration(config.RandomInterval))
	client.randomReset = make(chan bool)
	go client.randomFact()

	return client, nil
}

func (p *Pail) Open() error {
	return p.session.Open()
}

func (p *Pail) Close() {
	p.session.Close()
}

func (p *Pail) Reset() {
	p.randomReset <- true
}

func (p *Pail) Say(chanID, msg string) {
	if strings.TrimSpace(msg) != "" {
		log.Debug("Say: ", msg)
		p.session.ChannelMessageSend(chanID, msg)
		p.Reset()
	}
}

func (p *Pail) randomFact() {
	saidRandomFact := true
	for {
		select {
		case <-p.randomTicker.C:
			// TODO: This could be smarter and keep track of a seperate ticker per channel
			if !saidRandomFact {
				for _, chanID := range p.config.RandomChannels {
					fact, err := getRandomFact(p.db)
					if err != nil {
						p.randomReset <- true
						continue
					}
					log.Debug(fmt.Sprintf("Random fact: %+v", fact))
					if fact != nil {
						p.lastFact = fact
						reply, err := fact.handle()
						if err != nil {
							log.Error(err)
							p.randomReset <- true
							continue
						}
						log.Debug("Random fact reply: ", reply)
						p.session.ChannelMessageSend(strconv.Itoa(chanID), reply)
						saidRandomFact = true
					}
				}
			}
		case <-p.randomReset:
			p.randomTicker.Reset(time.Minute * time.Duration(p.config.RandomInterval))
			saidRandomFact = false
		}
	}
}
