package pubsub

import (
	"time"
)

type Level int

// stolen from ruby logger.rb
var (
	LevelDebug   = 0
	LevelInfo    = 1
	LevelWarn    = 2
	LevelError   = 3
	LevelFatal   = 4
	LevelUnknwon = 5
)

type Message struct {
	CreatedAt time.Time
	Key       string
	Message   string
	Duration  time.Duration
	Level     Level
	Payload   interface{}
}
