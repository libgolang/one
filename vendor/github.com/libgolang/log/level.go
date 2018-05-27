package log

import (
	"encoding/json"
	"strings"
)

const (
	// DEBUG debug level
	DEBUG Level = 40

	// INFO debug level
	INFO Level = 30

	// WARN warn level
	WARN Level = 20

	// ERROR error level
	ERROR Level = 10

	// OTHER empty value
	OTHER Level = 0
)

// Level Logging Level
type Level int

func (l Level) String() string {
	return levelToStr(l)
}

// MarshalJSON json serializer
func (l *Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(levelToStr(*l))
}

// UnmarshalJSON json unserializer
func (l *Level) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	*l = strToLevel(str)
	return err
}

func strToLevel(str string) Level {
	str = strings.ToUpper(str)
	var level Level
	switch str {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	default:
		level = OTHER
	}
	return level
}

func levelToStr(level Level) string {
	var str string

	switch level {
	case DEBUG:
		str = "DEBUG"
	case INFO:
		str = "INFO"
	case WARN:
		str = "WARN"
	case ERROR:
		str = "ERROR"
	default:
		str = "OTHER"
	}

	return str
}
