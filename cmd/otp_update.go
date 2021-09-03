package main

import (
	"github.com/jtaylorcpp/sts"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(rolesCmd)
	updateCmd.AddCommand(authorizersCmd)
	updateCmd.AddCommand(deviceCmd)

	rolesCmd.PersistentFlags().StringArrayVar(&stsRoles, "roles", []string{}, "AWS Role names to add to issue/account STS")

	authorizersCmd.PersistentFlags().StringArrayVar(&stsSecondaryAuthorizers, "secondary-authorizers", []string{}, "Secondary Account Names to do multi-person auth")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update various TOTP attributes",
}

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Update the roles for a given issuer/account name",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Debug().Str("issuer", sts.GetStringFlag("issuer")).
			Str("account_name", sts.GetStringFlag("account_name")).
			Msg("Updating roles")
		err := sts.UpdateTOTPEntryRoles(stsTableName, stsIssuer, stsAccountName, stsRoles, false)
		if err != nil {
			logger.Err(err).Msg("Error updating roles")
		}
	},
}

var authorizersCmd = &cobra.Command{
	Use:   "authorizers",
	Short: "Update the authorizers for a given issuer/account name",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Debug().Str("issuer", sts.GetStringFlag("issuer")).
			Str("account_name", sts.GetStringFlag("account_name")).
			Msg("Updating secondary authorizers")
		err := sts.UpdateTOTPEntrySecondaryAuthorizations(stsTableName, stsIssuer, stsAccountName, stsSecondaryAuthorizers, false)
		if err != nil {
			logger.Err(err).Msg("Error updating secondary authorizers")
		}
	},
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Update the device used for a given issuer/account name",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Debug().Str("issuer", sts.GetStringFlag("issuer")).
			Str("account_name", sts.GetStringFlag("account_name")).
			Msg("Updating device")
		key, err := sts.GenerateNewTOTP(stsIssuer, stsAccountName)
		if err != nil {
			logger.Err(err).Msg("Error generating new OTP key")
		}

		err = sts.DisplayTOTPQR(key)
		if err != nil {
			logger.Err(err).Msg("Error displaying new OTP key")
		}

		score := 0
		for score <= 1 {
			logger.Debug().Int("score", score).Msg("Current valid messages")
			logger.Print("Enter code to validate OTP code")
			var valid bool
			valid, err = sts.ValidateTOTPFromCLI(key)
			if err != nil {
				logger.Err(err).Msg("Error valdating OTP code")
			}
			if !valid {
				logger.Debug().Msg("Invalid OTP code given")
			} else {
				score += 1
				logger.Debug().Msg("Code accepted")
			}
		}

		err = sts.UpdateTOTPEntryMFADevice(stsTableName, stsIssuer, stsAccountName, key)
		if err != nil {
			logger.Err(err).Msg("Error updating device")
		}
	},
}
