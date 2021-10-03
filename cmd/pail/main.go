package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"time"

	"github.com/FryDay/pail"
)

var token = os.Getenv("PAIL_TOKEN")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	user, _ := user.Current()
	bot, err := pail.NewClient(token, user.HomeDir+"/.config/pail/pail.db")
	if err != nil {
		log.Fatalln(err)
	}

	err = bot.Open()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Pail is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc

	bot.Close()
}
