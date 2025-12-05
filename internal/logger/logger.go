package logger

import (
	"go.uber.org/zap"
	"log"
)

var Log *zap.Logger

func Init() {
	var err error
	Log, err = zap.NewDevelopment()
	if err != nil {
		log.Fatal("can't initialize zap logger: %v", err)
	}
}

func Sync() {
	_ = Log.Sync()
}
