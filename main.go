package main

import (
	"discord-verification/app"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		err = fmt.Errorf("an error has occurred while reading the environment variables: %v", err)
		panic(err)
	}

	config := new(app.Config)
	config.Load()

	bot := new(app.DiscordBot)
	bot.Start(*config)

	fmt.Println("Waiting on interruption...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
