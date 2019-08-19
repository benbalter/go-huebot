package hue

import (
	"fmt"
	"log"
	"net/http"

	"github.com/creasty/defaults"
)

const red = 0
const yellow = 12750

type State struct {
	On    bool   `default:"true" json:"on"`
	Bri   int    `default:"255" json:"bri"`
	Alert string `default:"none" json:"alert"`
	Sat   int    `default:"255" json:"sat"`
	Hue   int    `json:"hue,omitempty"`
}

func pathForLight(light int) string {
	return fmt.Sprintf("/bridge/%s/lights/%d/state", Username, light)
}

func stateForStatus(status string) State {
	state := State{}

	switch status {
	case "none":
		state.Sat = 0
	case "minor":
		state.Hue = yellow
		state.Alert = "lselect"
	case "major", "critical":
		state.Hue = red
		state.Alert = "lselect"

	}
	err := defaults.Set(&state)

	if err != nil {
		log.Fatal(err)
	}

	return state
}

// SetColor sets a light to the proper color for the given status
func SetColor(light int, status string) {
	state := stateForStatus(status)
	path := pathForLight(light)
	MakeRequest(path, http.MethodPut, &state, nil)
}
