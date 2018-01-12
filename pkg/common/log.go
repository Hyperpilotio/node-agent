package common

import (
	log "github.com/sirupsen/logrus"
)

func GetLevel(s string) log.Level {
	switch s {
	case "1":
		return log.DebugLevel
	case "2":
		return log.InfoLevel
	case "3":
		return log.WarnLevel
	case "4":
		return log.ErrorLevel
	case "5":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
