package hue_test

import (
	"testing"

	"github.com/benbalter/go-huebot/pkg/hue"
	"github.com/h2non/gock"
)

func TestSetColor(t *testing.T) {
	defer gock.Off()

	state := hue.State{true, 255, "none", 255, 0}
	gock.New("https:///api.meethue.com").Put("/bridge/(.*)/lights/1/state").
		JSON(state).Reply(200)

	hue.SetColor(1, "good")
}
