package main

import (
	"github.com/ArcaneCrowA/go-url-shortener/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	app.Start()
}
