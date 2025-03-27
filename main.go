package main

import (
	"stickers-bot/BotsController/DiscordBot"
)

func main() {
	DiscordBot.Start()
	<-make(chan struct{})
}