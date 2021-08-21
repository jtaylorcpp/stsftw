package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jtaylorcpp/sts"
	"github.com/spf13/cobra"
)

var primaryTOTPCode string
var secondaryTOTPCode string

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.PersistentFlags().StringVar(&stsRole, "role", "", "AWS Role name to get AWS STS creds for from STS")
	getCmd.PersistentFlags().StringVar(&stsSecondaryAuthorizer, "secondary-authorizer", "", "Seconadry Authorizer for Multi-Party Auth")

	getCmd.PersistentFlags().StringVar(&primaryTOTPCode, "totp-code", "", "TOTP code from an enrolled device.")
	getCmd.PersistentFlags().StringVar(&secondaryTOTPCode, "secondary-totp-code", "", "TOTP code from an enrolled device (secondary).")
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get AWS STS credentials from STS",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Info().Msg("Getting STS credentials")
		challenge := sts.TOTPChallenge{
			Issuer:               sts.GetStringFlag("issuer"),
			AccountName:          sts.GetStringFlag("account_name"),
			TOTPCode:             primaryTOTPCode,
			SecondaryAccountName: sts.GetStringFlag("secondary_authorizer"),
			SecondaryTOTPCode:    secondaryTOTPCode,
			Role:                 sts.GetStringFlag("role"),
		}

		logger.Info().Msg(fmt.Sprintf("Challenge for STS: %#v\n", challenge))

		apiTimeout := time.Duration(time.Minute)
		apiclient := http.Client{
			Timeout: apiTimeout,
		}

		jsonBody, jsonErr := json.Marshal(&challenge)
		if jsonErr != nil {
			logger.Fatal().Err(jsonErr).Msg("Failed to create json event to get STS creds")
		}

		postRequest, requestErr := http.NewRequest("POST", sts.GetStringFlag("endpoint"), bytes.NewBuffer(jsonBody))
		if requestErr != nil {
			logger.Fatal().Err(requestErr).Msg("Error creating POST request to get STS creds")
		}

		postRequest.Header.Set("Content-Type", "application/json")

		response, responseErr := apiclient.Do(postRequest)
		if responseErr != nil {
			logger.Fatal().Err(responseErr).Msg("Error in response from POST to STS API")
		}

		defer response.Body.Close()

		body, bodyErr := ioutil.ReadAll(response.Body)
		if bodyErr != nil {
			logger.Fatal().Err(bodyErr).Msg("Unable to open response body")
		}

		creds := sts.STSCredentials{}
		unmarshallErr := json.Unmarshal(body, &creds)
		if unmarshallErr != nil {
			logger.Fatal().Err(unmarshallErr).Msg("Unable to unmarshal JSON response")
		}

		fmt.Printf("AWS_ACCESS_KEY_ID=%s\nAWS_SECRET_ACCESS_KEY=%s\nAWS_SESSION_TOKEN=%s\n",
			creds.AccessKeyId,
			creds.SecretAccessKey,
			creds.SessionToken)
	},
}
