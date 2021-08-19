package sts

import (
	"bufio"
	"errors"
	"os"

	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
)

func GenerateNewTOTP(issuer, accountName string) (*otp.Key, error) {
	logrus.Infof("Creating TOTP Key for: %s, %s\n", issuer, accountName)
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
}

func DisplayTOTPQR(key *otp.Key) error {
	qrterminal.GenerateWithConfig(key.URL(), qrterminal.Config{
		Level:     qrterminal.H,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	})
	return nil
}

func ValidateTOTPKey(code string, key *otp.Key) (bool, error) {
	valid := totp.Validate(code, key.Secret())
	if valid {
		return true, nil
	} else {
		return false, nil
	}
}

func ValidateTOTPURL(code, url string) (bool, error) {
	key, keyErr := otp.NewKeyFromURL(url)
	if keyErr != nil {
		return false, keyErr
	}

	return ValidateTOTPKey(code, key)
}

func ValidateTOTPFromCLI(key *otp.Key) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	logrus.Print("Enter Passcode: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		logrus.Errorln(err.Error())
		return false, err
	}
	return ValidateTOTPKey(text, key)
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
	logrus.Infof("Validating role that %s is in set %v\n", t.Role, entry.Roles)
	if len(entry.Roles) == 0 {
		return false
	}

	for _, role := range entry.Roles {
		if role == t.Role {
			return true
		}
	}

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
	for _, pair := range pairs {
		valid, err := ValidateTOTPURL(pair.Code, pair.URL)
		if err != nil {
			return err
		}

		if !valid {
			return errors.New("Invalid passcode")
		}
	}

	return nil
}
