package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jtaylorcpp/sts"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	logger := sts.GetLogger()
	// check if body is the struct we expect
	challenge := sts.TOTPChallenge{}
	jsonErr := json.Unmarshal([]byte(request.Body), &challenge)
	// if there are any errors or there is no issuer/passcode reject
	if jsonErr != nil || challenge.Issuer == "" || challenge.AccountName == "" || challenge.Role == "" {
		if jsonErr != nil {
			logger.Err(jsonErr).Msg("Unable to unmarshall client request")
		}
		logger.Info().Msg("Blackholing response")
		// blackhole
		return returnBlackhole()
	}

	challengeSet := []sts.ChallengePair{}
	logger.WithChallenge(challenge)
	logger.Info().Str("table", sts.GetStringFlag("table_name")).Msg("getting entry from table")

	firstEntry, getFirstEntryErr := sts.GetTOTPEntry(sts.GetStringFlag("table_name"), challenge.Issuer, challenge.AccountName)
	if getFirstEntryErr != nil {
		logger.Err(getFirstEntryErr).Msg("Unable to find user")
		return returnError(errors.New("Unable to find user"))
	}

	logger.WithChallenge(challenge)
	logger.WithEntry(firstEntry)
	logger.Info().Msg("Validating challenge role against entry")
	if challenge.ValidateRole(firstEntry) == false {
		return returnError(errors.New("Invalid role"))
	}

	challengeSet = append(challengeSet, sts.ChallengePair{challenge.TOTPCode, firstEntry.URL})

	if challenge.NeedsSecondaryValidation(firstEntry) {
		secondEntry, getSecondEntryErr := sts.GetTOTPEntry(sts.GetStringFlag("table_name"), challenge.Issuer, challenge.SecondaryAccountName)
		if getSecondEntryErr != nil {
			logger.WithChallenge(challenge)
			logger.Err(getSecondEntryErr).Msg("Error getting secondary auth")
			return returnError(errors.New("Unable to find user(secondary)"))
		}

		challengeSet = append(challengeSet, sts.ChallengePair{challenge.SecondaryTOTPCode, secondEntry.URL})
	}

	challengeErr := sts.ValidateChallengeSet(challengeSet)
	if challengeErr != nil {
		logger.WithChallenge(challenge)
		logger.WithEntry(firstEntry)
		logger.Err(challengeErr).Msg("Unable to validate both codes")
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
	return events.ALBTargetGroupResponse{Body: "", StatusCode: 500, StatusDescription: "5 Internal Server Error", IsBase64Encoded: false, Headers: map[string]string{}}, err
}

func returnCreds(challenge sts.TOTPChallenge) (events.ALBTargetGroupResponse, error) {
	logger := sts.GetLogger()
	// make call to audit notifier first
	publishErr := sts.PublishAuditEvent(challenge.Issuer, challenge.AccountName, challenge.Role)
	if publishErr != nil {
		logger.WithChallenge(challenge)
		logger.Err(publishErr).Msg("Unable to publish cred minting notification")
		return returnError(errors.New("Unable to publish audit event"))
	}

	// then mint creds
	creds, err := sts.GetCredentials(challenge.Issuer, challenge.AccountName, challenge.Role)
	if err != nil {
		logger.WithChallenge(challenge)
		logger.Err(err).Msg("Uable to mint credentials")
		return returnError(err)
	}

	jsonBytes, jsonErr := json.Marshal(&creds)
	if jsonErr != nil {
		logger.WithChallenge(challenge)
		logger.Err(jsonErr).Msg("Unable to marshall credentials")
		return returnError(jsonErr)
	}

	return events.ALBTargetGroupResponse{Body: string(jsonBytes), StatusCode: 200, StatusDescription: "200 Ok", IsBase64Encoded: false, Headers: map[string]string{}}, nil
}
