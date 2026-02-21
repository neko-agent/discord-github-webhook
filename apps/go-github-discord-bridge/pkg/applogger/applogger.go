package applogger

import (
	"dizzy/logger"
	"dizzy/logger/strategies"
)

var Log logger.Logger

func Init(environment string) {
	if environment == "production" {
		Log = strategies.NewZapMust(strategies.ZapOptions{
			ServiceName: "github-discord-bridge",
			Level:       strategies.InfoLevel,
		})
	} else {
		Log = strategies.NewZapMust(strategies.ZapOptions{
			ServiceName: "github-discord-bridge",
			IsPretty:    true,
			Level:       strategies.DebugLevel,
		})
	}
}
