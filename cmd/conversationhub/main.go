package main

import (
	"log"

	"github.com/joeshaw/envdecode"
)

// we need to
// emit registratrion event on boot up
// listen for messages
// print them, reply to event.source maybe?
// randomly emit $ACTOR_MESSAGE

type Config struct {
	CONVERSATION_ string        `env:"SERVER_HOSTNAME,default=localhost"`
	Port          uint16        `env:"SERVER_PORT,default=8080"`
	Timeout       time.Duration `env:"TIMEOUT,default=1m,strict"`
}

func main() {
	username := viper.GetString("USERNAME")
	log.Printf("username %s", username)

}
