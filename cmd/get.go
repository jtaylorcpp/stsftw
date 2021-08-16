package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jtaylorcpp/sts"
	"github.com/sirupsen/logrus"
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
		logrus.Infoln("Getting STS creds")
		challenge := sts.TOTPChallenge{
			Issuer:               sts.GetStringFlag("issuer"),
			AccountName:          sts.GetStringFlag("account_name"),
			TOTPCode:             primaryTOTPCode,
			SecondaryAccountName: sts.GetStringFlag("secondary_authorizer"),
			SecondaryTOTPCode:    secondaryTOTPCode,
			Role:                 sts.GetStringFlag("role"),
		}

		logrus.Infof("Challenge for STS: %#v\n", challenge)

		apiTimeout := time.Duration(time.Minute)
		apiclient := http.Client{
			Timeout: apiTimeout,
		}

		jsonBody, jsonErr := json.Marshal(&challenge)
		if jsonErr != nil {
			logrus.Fatalln(jsonErr.Error())
		}

		postRequest, requestErr := http.NewRequest("POST", sts.GetStringFlag("endpoint"), bytes.NewBuffer(jsonBody))
		if requestErr != nil {
			logrus.Fatalln(requestErr.Error())
		}

		postRequest.Header.Set("Content-Type", "application/json")

		response, responseErr := apiclient.Do(postRequest)
		if responseErr != nil {
			logrus.Fatalln(responseErr.Error())
		}

		defer response.Body.Close()

		body, bodyErr := ioutil.ReadAll(response.Body)
		if bodyErr != nil {
			logrus.Fatalln(bodyErr.Error())
		}

		creds := sts.STSCredentials{}
		unmarshallErr := json.Unmarshal(body, &creds)
		if unmarshallErr != nil {
			logrus.Fatalln(unmarshallErr.Error())
		}

		logrus.Infof("creds %#v\n", creds)
	},
}