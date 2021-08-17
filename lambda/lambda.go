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

	challengeSet := []sts.ChallengePair{}

	firstEntry, getFirstEntryErr := sts.GetTOTPEntry(sts.GetStringFlag("table_name"), challenge.Issuer, challenge.AccountName)
	if getFirstEntryErr != nil {
		logrus.Errorln(getFirstEntryErr.Error())
		return returnError(errors.New("Unable to find user"))
	}

	if !challenge.ValidateRole(firstEntry) {
		returnError(errors.New("Invalid role"))
	}

	challengeSet = append(challengeSet, sts.ChallengePair{challenge.TOTPCode, firstEntry.URL})

	if challenge.NeedsSecondaryValidation(firstEntry) {
		secondEntry, getSecondEntryErr := sts.GetTOTPEntry(sts.GetStringFlag("table_name"), challenge.Issuer, challenge.SecondaryAccountName)
		if getSecondEntryErr != nil {
			logrus.Errorln(getSecondEntryErr.Error())
			return returnError(errors.New("Unable to find user(secondary)"))
		}

		challengeSet = append(challengeSet, sts.ChallengePair{challenge.SecondaryTOTPCode, secondEntry.URL})
	}

	challengeErr := sts.ValidateChallengeSet(challengeSet)
	if challengeErr != nil {
		logrus.Errorln(challengeErr.Error())
		return returnError(errors.New("Unable to validate codes"))
	}

	return returnCreds(challenge)
	//return events.ALBTargetGroupResponse{Body: request.Body, StatusCode: 200, StatusDescription: "200 OK", IsBase64Encoded: false, Headers: map[string]string{}}, nil
}

func returnBlackhole() (events.ALBTargetGroupResponse, error) {
	return events.ALBTargetGroupResponse{Body: "", StatusCode: 308, StatusDescription: "308 Permanent Redirect", IsBase64Encoded: false, Headers: map[string]string{
		"Location": "https://science.nasa.gov/astrophysics/focus-areas/black-holes",
	}}, errors.New("Unknown client")
}

func returnError(err error) (events.ALBTargetGroupResponse, error) {
	return events.ALBTargetGroupResponse{Body: "", StatusCode: 511, StatusDescription: "511 Network Authentication Required", IsBase64Encoded: false, Headers: map[string]string{}}, err
}

func returnCreds(challenge sts.TOTPChallenge) (events.ALBTargetGroupResponse, error) {
	// make call to audit notifier first
	publishErr := sts.PublishAuditEvent(challenge.Issuer, challenge.AccountName, challenge.Role)
	if publishErr != nil {
		logrus.Errorln(publishErr.Error())
		return returnError(errors.New("Unable to publish audit event"))
	}

	// then mint creds
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
