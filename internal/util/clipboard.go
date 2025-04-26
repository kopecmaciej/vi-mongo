package util

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/rs/zerolog/log"
)

func GetClipboard() (func(string), func() string) {
	cpFunc := func(text string) {
		err := clipboard.WriteAll(text)
		if err != nil {
			log.Error().Err(err).Msg("Error writing to clipboard")
		}
	}
	pasteFunc := func() string {
		text, err := clipboard.ReadAll()
		if err != nil {
			log.Error().Err(err).Msg("Error reading from clipboard")
			return ""
		}
		return strings.TrimSpace(text)
	}

	return cpFunc, pasteFunc
}
