package poopchat

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
)

const randomAPIURL = "https://random-word-api.herokuapp.com/word?number=1&swear=1"

type username string

var usernameValidation = regexp.MustCompile(`^[a-zA-Z0-9]+(?:-[a-zA-Z0-9]+)*$`)

func (u username) valid() bool {
	return usernameValidation.MatchString(string(u))
}

func (u username) string() string {
	return string(u)
}

func (u username) bytes() []byte {
	return []byte(u)
}

func getRandomUsername(ctx context.Context) username {
	logger := ctx.Value(SessionLoggerKey).(*log.Entry)

	resp, err := http.Get(randomAPIURL)
	if err != nil {
		logger.WithContext(ctx).WithError(err).Error("Failed to set random username")
		return "fallbackUsername"
	}

	var userNames []string
	if err = json.NewDecoder(resp.Body).Decode(&userNames); err != nil {
		logger.WithContext(ctx).WithError(err).Error("Failed to decode random username")
		return "fallbackUsername"
	}

	return username(userNames[0])
}
