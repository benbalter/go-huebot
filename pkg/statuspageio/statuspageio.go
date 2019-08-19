package statuspageio

import (
	"log"

	"github.com/dghubble/sling"
	_ "github.com/joho/godotenv/autoload"
)

const endpoint = "https://www.githubstatus.com/api/v2/status.json"

type Page struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	TimeZone  string `json:"time_zone"`
	UpdatedAt string `json:"updated_at"`
}

type Status struct {
	Indicator   string `json:"indicator"`
	Description string `json:"description"`
}

type Response struct {
	Page   Page   `json:"page"`
	Status Status `json:"status"`
}

// GetStatus returns the current status of GitHub.com
func GetStatus() Status {
	resp := Response{}
	_, err := sling.New().Get(endpoint).ReceiveSuccess(&resp)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Current status: \"%s\" (%s)", resp.Status.Indicator, resp.Status.Description)

	return resp.Status
}
