package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/benbalter/go-huebot/pkg/hue"
	"github.com/benbalter/go-huebot/pkg/statuspageio"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	auth := flag.Bool("auth", false, "Init OAuth flow")
	reset := flag.Bool("reset", false, "Reset creds")

	flag.Parse()

	if *auth {
		hue.AuthDance()
	}

	if *reset {
		hue.Reset()
	}

	lightID, err := strconv.Atoi(os.Getenv("HUE_LIGHT_ID"))
	if err != nil {
		log.Fatal(err)
	}

	status := statuspageio.GetStatus()
	hue.SetColor(lightID, status.Indicator)
}
