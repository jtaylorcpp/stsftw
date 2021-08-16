package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jtaylorcpp/sts"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	// check if body is the struct we expect
	challenge := sts.TOTPChallenge{}
	jsonErr := json.Unmarshal([]byte(request.Body), &challenge)
	// if there are any errors or there is no issuer/passcode reject
	if jsonErr != nil || challenge.Issuer == "" || challenge.AccountName == "" || challenge.Role == "" {
		if jsonErr != nil {
			logrus.Errorln(jsonErr.Error())
		}
		logrus.Println("Blackholing response")
		// blackhole
		return returnBlackhole()
	}

	// no errors and data so lets go
	firstEntry, getFirstEntryErr := sts.GetTOTPEntry(sts.GetStringFlag("table_name"), challenge.Issuer, challenge.AccountName)
	if getFirstEntryErr != nil {
		if !challenge.ValidateRole(firstEntry) {
			logrus.Errorf("Role %s is not in entry %v\n", challenge.Role, firstEntry)
			return returnError(errors.New("Invalid role"))
		}

		if primaryErr := challenge.ValidatePrimaryChallenge(firstEntry); primaryErr == nil {
			if challenge.NeedsSecondaryValidation(firstEntry) {
				if secondaryErr := challenge.ValidateSecondaryChallenge(sts.GetStringFlag("table_name"), firstEntry); secondaryErr != nil {
					// error on validating secondary
					logrus.Errorln(secondaryErr.Error(), " - error validating second code")
					return returnError(secondaryErr)
				} else {
					// valid code and return creds
					return returnCreds(challenge)
				}
			} else {
				// valid code and return creds
				return returnCreds(challenge)
			}
		} else {
			// error out for not validating primary code
			logrus.Errorln(primaryErr.Error(), " - primary code is not valid")
			return returnError(primaryErr)
		}
	} else {
		// error out for not being able to get totp info
		logrus.Errorln(getFirstEntryErr.Error())
		// add error return
		return returnError(getFirstEntryErr)
	}

	//return events.ALBTargetGroupResponse{Body: request.Body, StatusCode: 200, StatusDescription: "200 OK", IsBase64Encoded: false, Headers: map[string]string{}}, nil
}

func returnBlackhole() (events.ALBTargetGroupResponse, error) {
	return events.ALBTargetGroupResponse{Body: "", StatusCode: 308, StatusDescription: "308 Permanent Redirect", IsBase64Encoded: false, Headers: map[string]string{
		"Location": "https://science.nasa.gov/astrophysics/focus-areas/black-holes",
	}}, nil
}

func returnError(err error) (events.ALBTargetGroupResponse, error) {
	return events.ALBTargetGroupResponse{Body: "", StatusCode: 511, StatusDescription: "511 Network Authentication Required", IsBase64Encoded: false, Headers: map[string]string{}}, err
}

func returnCreds(challenge sts.TOTPChallenge) (events.ALBTargetGroupResponse, error) {
	creds, err := sts.GetCredentials(challenge.Issuer, challenge.AccountName, challenge.Role)
	if err != nil {
		logrus.Errorln(err.Error())
		return returnError(err)
	}

	jsonBytes, jsonErr := json.Marshal(&creds)
	if jsonErr != nil {
		logrus.Errorln(jsonErr.Error())
		return returnError(jsonErr)
	}

	return events.ALBTargetGroupResponse{Body: string(jsonBytes), StatusCode: 200, StatusDescription: "200 Ok", IsBase64Encoded: false, Headers: map[string]string{}}, nil
}
