package main

import (
	"fmt"

	"github.com/jtaylorcpp/sts"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(enrollCmd)

	enrollCmd.PersistentFlags().StringArrayVar(&stsRoles, "roles", []string{}, "AWS Role names to add to issue/account STS")

	enrollCmd.PersistentFlags().StringArrayVar(&stsSecondaryAuthorizers, "secondary-authorizers", []string{}, "Secondary Account Names to do multi-person auth")

}

var enrollCmd = &cobra.Command{
	Use:   "enroll",
	Short: "Enroll your device in STS",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Info().Msgf("Enrolling new device")
		key, err := sts.GenerateNewTOTP(sts.GetStringFlag("issuer"), sts.GetStringFlag("account_name"))
		if err != nil {
			logger.Error().Err(err)
		}

		err = sts.DisplayTOTPQR(key)
		if err != nil {
			logger.Error().Err(err)
		}

		score := 0
		for score <= 1 {
			fmt.Println("Validate TOTP enrollment")
			var valid bool
			valid, err = sts.ValidateTOTPFromCLI(key)

			if err != nil {
				logger.Error().Err(err)
			}
			if !valid {
				logger.Error().Err(err)
			} else {
				score += 1
				fmt.Println("Code accepted")
			}
		}

		fmt.Println("TOTP enrolled and verified")

		tableEntry, entryErr := sts.NewTOTPEntry(key)
		if entryErr != nil {
			logrus.Errorln(entryErr.Error())
		}

		tableEntry.SecondaryAuthorization = sts.GetStringArrayFlag("secondary_authorizers")
		tableEntry.Roles = sts.GetStringArrayFlag("roles")
		logrus.Infof("Adding entry to DynamoDB: %#v\n", tableEntry)
		sts.AddTOTPEntryToTable(sts.GetStringFlag("table_name"), tableEntry)
	},
}
