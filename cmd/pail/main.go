package main

import (
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/FryDay/pail"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.MkdirAll("logs", 0700)
	}

	logFile, err := os.OpenFile("logs/pail.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	confDir := filepath.Join(os.Getenv("HOME"), ".config/pail")
	if err := os.MkdirAll(confDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	var conf pail.Config
	confPath := filepath.Join(confDir, "pail.toml")
	confFile, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := toml.Decode(string(confFile), &conf); err != nil {
		log.Fatalln(err)
	}
	dbPath := filepath.Join(confDir, "pail.db")
	log.SetLevel(log.InfoLevel)
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
	}

	pail, err := pail.NewPail(&conf, dbPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = pail.Open()
	if err != nil {
		log.Fatalln(err)
	}

	log.Info("Pail is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc

	pail.Close()
}
