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

func ValidateTOTP(key *otp.Key) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	logrus.Print("Enter Passcode: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		logrus.Errorln(err.Error())
		return false, err
	}
	valid := totp.Validate(text, key.Secret())
	if valid {
		return true, nil
	} else {
		return false, nil
	}
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
	for _, role := range entry.Roles {
		if role != t.Role {
			return true
		}
	}

	return false
}

func (t TOTPChallenge) ValidatePrimaryChallenge(entry TOTPEntry) error {
	key, keyErr := EntryToTOTP(entry)
	if keyErr != nil {
		logrus.Errorln(keyErr.Error())
		return keyErr
	}

	if !totp.Validate(t.TOTPCode, key.Secret()) {
		return errors.New("Invalid Code")
	}

	return nil
}

func (t TOTPChallenge) NeedsSecondaryValidation(entry TOTPEntry) bool {
	if len(entry.SecondaryAuthorization) > 0 {
		return true
	}

	return false
}

func (t TOTPChallenge) ValidateSecondaryChallenge(tableName string, entry TOTPEntry) error {
	correctSecondary := false
	for _, secondary := range entry.SecondaryAuthorization {
		if secondary == t.SecondaryAccountName {
			correctSecondary = true
			break
		}
	}

	if correctSecondary {
		secondaryEntry, getErr := GetTOTPEntry(tableName, t.Issuer, t.SecondaryAccountName)
		if getErr != nil {
			logrus.Errorln(getErr.Error())
			return getErr
		}

		secondaryKey, keyErr := EntryToTOTP(secondaryEntry)
		if keyErr != nil {
			logrus.Errorln(keyErr.Error())
			return keyErr
		}

		if totp.Validate(t.SecondaryTOTPCode, secondaryKey.Secret()) {
			return nil
		} else {
			return errors.New("Secondary code is invalid")
		}
	} else {
		return errors.New("Incorrect secondary account name")
	}
}
