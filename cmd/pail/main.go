package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/FryDay/pail"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	confDir := filepath.Join(os.Getenv("HOME"), ".config/pail")
	if err := os.MkdirAll(confDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	var conf pail.Config
	confPath := filepath.Join(confDir, "pail.toml")
	confFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := toml.Decode(string(confFile), &conf); err != nil {
		log.Fatalln(err)
	}
	dbPath := filepath.Join(confDir, "pail.db")

	pail, err := pail.NewPail(&conf, dbPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = pail.Open()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Pail is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc

	pail.Close()
}
