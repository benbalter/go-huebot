package statuspageio_test

import (
	"testing"

	"github.com/benbalter/go-huebot/pkg/statuspageio"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	defer gock.Off()
	assert := assert.New(t)

	status := statuspageio.Status{"foo", "bar"}
	response := statuspageio.Response{statuspageio.Page{}, status}

	gock.New("https://www.githubstatus.com/").
		Get("api/v2/status.json").Reply(200).
		JSON(response)

	result := statuspageio.GetStatus()
	assert.Equal(status, result, "Expected status to match")
}
