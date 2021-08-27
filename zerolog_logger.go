package sts

import (
	"os"

	"github.com/pquerna/otp"
	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
	entry     chan TOTPEntry
	challenge chan TOTPChallenge
	key       chan *otp.Key
}

func (l Logger) WithEntry(entry TOTPEntry) {
	if len(l.entry) == 0 {
		l.entry <- entry
	}
}

func (l Logger) WithChallenge(challenge TOTPChallenge) {
	if len(l.challenge) == 0 {
		l.challenge <- challenge
	}

}

func (l Logger) WithKey(key *otp.Key) {
	if len(l.key) == 0 {
		l.key <- key
	}

}

type annotationHook struct {
	entry     chan TOTPEntry
	challenge chan TOTPChallenge
	key       chan *otp.Key
}

func (a annotationHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	if len(a.entry) > 0 {
		entry := <-a.entry

		e.Str("entry_issuer", entry.Issuer).
			Str("entry_account_name", entry.AccountName).
			Strs("entry_roles", entry.Roles).
			Strs("entry_secondaries", entry.SecondaryAuthorization)
	}

	if len(a.challenge) > 0 {
		challenge := <-a.challenge

		e.Str("challenge_issuer", challenge.Issuer).
			Str("challenge_account_name", challenge.AccountName).
			Str("challenge_role", challenge.Role).
			Str("challenge_secondary", challenge.SecondaryAccountName)
	}

	if len(a.key) > 0 {
		key := <-a.key
		e.Str("key_issuer", key.Issuer()).
			Str("key_account_name", key.AccountName())
	}
}

var logger Logger
var logLevelNeedsSetting bool = true

func init() {
	entryChan := make(chan TOTPEntry, 1)
	challengeChan := make(chan TOTPChallenge, 1)
	keyChan := make(chan *otp.Key, 1)
	hook := annotationHook{
		entryChan,
		challengeChan,
		keyChan,
	}

	logger = Logger{
		zerolog.New(os.Stderr).Hook(hook).With().Timestamp().Logger(),
		entryChan,
		challengeChan,
		keyChan,
	}
}

func GetLogger() Logger {
	if logLevelNeedsSetting {
		SetLogLevel()
		logLevelNeedsSetting = false
	}

	return logger
}

func SetLogLevel() {
	// set log level
	switch GetStringFlag("log_level") {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
}
