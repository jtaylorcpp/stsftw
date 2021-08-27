package sts

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateNewTOTP(issuer, accountName string) (*otp.Key, error) {
	logger := GetLogger()
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Msg("Creating new TOTP entry")
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
}

func DisplayTOTPQR(key *otp.Key) error {
	logger := GetLogger()
	logger.WithKey(key)
	logger.Trace().Msg("Displaying QR code to enroll MFA device")
	qrterminal.GenerateWithConfig(key.URL(), qrterminal.Config{
		Level:     qrterminal.H,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	})
	return nil
}

func ValidateTOTPKey(code string, key *otp.Key) bool {
	logger := GetLogger()
	logger.WithKey(key)
	logger.Trace().Msg("Validating TOTP code")
	valid := totp.Validate(code, key.Secret())
	if valid {
		logger.WithKey(key)
		logger.Trace().Msg("Valid TOTP code")
		return true
	} else {
		logger.WithKey(key)
		logger.Warn().Msg("Invalid TOTP code")
		return false
	}
}

func ValidateTOTPURL(code, url string) (bool, error) {
	logger := GetLogger()
	logger.Trace().Msg("Parsing TOTP key from URL")
	key, keyErr := otp.NewKeyFromURL(url)
	if keyErr != nil {
		logger.Error().Str("keyErr", keyErr.Error()).Msg("Error parsing TOTP key from URL")
		return false, keyErr
	}
	logger.WithKey(key)
	logger.Trace().Msg("Parsed TOTP key from URL")

	return ValidateTOTPKey(code, key), nil
}

func ValidateTOTPFromCLI(key *otp.Key) (bool, error) {
	logger := GetLogger()
	logger.WithKey(key)
	logger.Trace().Msg("Reading passcode in from CLI")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Passcode: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		logger.WithKey(key)
		logger.Error().Msg("Unable to read passcode from CLI")
		return false, err
	}
	logger.WithKey(key)
	logger.Trace().Msg("Passcode read from CLI")
	return ValidateTOTPKey(text, key), nil
}

type TOTPChallenge struct {
	Issuer               string
	AccountName          string
	TOTPCode             string
	SecondaryAccountName string
	SecondaryTOTPCode    string
	Role                 string
}

func (t TOTPChallenge) ValidateRole(entry TOTPEntry) bool {
	logger.Trace().Str("issuer", t.Issuer).Str("account_name", t.AccountName).Str("role", t.Role).Msg("Validating Role for TOTP Challenge")
	if len(entry.Roles) == 0 {
		return false
	}

	for _, role := range entry.Roles {
		if role == t.Role {
			logger.Trace().Str("issuer", t.Issuer).Str("account_name", t.AccountName).Str("role", t.Role).Msg("Role has been assigned")
			return true
		}
	}

	logger.Trace().Str("issuer", t.Issuer).Str("account_name", t.AccountName).Str("role", t.Role).Msg("Invalid/unassigned role")
	return false
}

func (t TOTPChallenge) NeedsSecondaryValidation(entry TOTPEntry) bool {
	if len(entry.SecondaryAuthorization) > 0 {
		return true
	}

	return false
}

type ChallengePair struct {
	Code string
	URL  string
}

func ValidateChallengeSet(pairs []ChallengePair) error {
	logger.Trace().Msg("Validating challenge set to authorize STS creds")
	for _, pair := range pairs {
		valid, err := ValidateTOTPURL(pair.Code, pair.URL)
		if err != nil {
			logger.Error().Str("error", err.Error()).Msg("Invalid challenge set and denying STS creds")
			return err
		}

		if !valid {
			logger.Warn().Str("warn", "invalid codes").Msg("Invalid challenge set and denying STS creds")
			return errors.New("Invalid passcode")
		}
	}

	logger.Trace().Msg("Validated challenge set to authorize STS creds")
	return nil
}
