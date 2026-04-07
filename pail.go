package pail

import (
	"fmt"
	"strings"
	"sync"
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
	lastFactMu   sync.Mutex
	randomTicker *time.Ticker
	randomReset  chan bool
	done         chan struct{}
	regexCache   map[bool][]*Regex
	regexCacheMu sync.RWMutex
}

func (p *Pail) setLastFact(f *Fact) {
	p.lastFactMu.Lock()
	defer p.lastFactMu.Unlock()
	p.lastFact = f
}

func (p *Pail) getLastFact() *Fact {
	p.lastFactMu.Lock()
	defer p.lastFactMu.Unlock()
	return p.lastFact
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
	client.done = make(chan struct{})
	if err := client.loadRegexCache(); err != nil {
		return nil, err
	}
	go client.randomFact()

	return client, nil
}

func (p *Pail) loadRegexCache() error {
	p.regexCacheMu.Lock()
	defer p.regexCacheMu.Unlock()
	mentionRegex, err := loadAllRegex(p.db, true)
	if err != nil {
		return err
	}
	normalRegex, err := loadAllRegex(p.db, false)
	if err != nil {
		return err
	}
	p.regexCache = map[bool][]*Regex{
		true:  mentionRegex,
		false: normalRegex,
	}
	return nil
}

func (p *Pail) getRegex(mention bool) []*Regex {
	p.regexCacheMu.RLock()
	defer p.regexCacheMu.RUnlock()
	return p.regexCache[mention]
}

func (p *Pail) Open() error {
	return p.session.Open()
}

func (p *Pail) Close() {
	close(p.done)
	p.randomTicker.Stop()
	p.session.Close()
}

func (p *Pail) Reset() {
	p.randomReset <- true
}

func (p *Pail) Say(chanID, msg string) {
	if strings.TrimSpace(msg) != "" {
		log.Debug("Say: ", msg)
		if _, err := p.session.ChannelMessageSend(chanID, msg); err != nil {
			log.Error("Failed to send message: ", err)
		}
		p.Reset()
	}
}

func (p *Pail) randomFact() {
	saidRandomFact := true
	for {
		select {
		case <-p.done:
			return
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
						p.setLastFact(fact)
						reply, err := fact.handle()
						if err != nil {
							log.Error(err)
							p.randomReset <- true
							continue
						}
						log.Debug("Random fact reply: ", reply)
						if _, err := p.session.ChannelMessageSend(chanID, reply); err != nil {
							log.Error("Failed to send random fact: ", err)
						}
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
