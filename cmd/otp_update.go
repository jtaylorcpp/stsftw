package main

import (
	"github.com/jtaylorcpp/sts"
	"github.com/sirupsen/logrus"
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
		logrus.Debugln("Updating roles for <issuer><account name>")
		err := sts.UpdateTOTPEntryRoles(stsTableName, stsIssuer, stsAccountName, stsRoles, false)
		if err != nil {
			logrus.Errorln(err.Error())
		}
	},
}

var authorizersCmd = &cobra.Command{
	Use:   "authorizers",
	Short: "Update the authorizers for a given issuer/account name",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debugln("Updating authorizers for <issuer><account name>")
		err := sts.UpdateTOTPEntrySecondaryAuthorizations(stsTableName, stsIssuer, stsAccountName, stsSecondaryAuthorizers, false)
		if err != nil {
			logrus.Errorln(err.Error())
		}
	},
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Update the device used for a given issuer/account name",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debugln("Updating device for <issuer><account name>")
		key, err := sts.GenerateNewTOTP(stsIssuer, stsAccountName)
		if err != nil {
			logrus.Errorln(err.Error())
		}

		err = sts.DisplayTOTPQR(key)
		if err != nil {
			logrus.Errorln(err.Error())
		}

		score := 0
		for score <= 1 {
			logrus.Debugf("%v successful attempts of 2 needed\n", score)
			logrus.Infoln("Validate TOTP enrollment")
			var valid bool
			valid, err = sts.ValidateTOTP(key)
			if err != nil {
				logrus.Errorln(err.Error())
			}
			if !valid {
				logrus.Errorln("incorrect code entered")
			} else {
				score += 1
				logrus.Infoln("Code accepted")
			}
		}

		err = sts.UpdateTOTPEntryMFADevice(stsTableName, stsIssuer, stsAccountName, key)
		if err != nil {
			logrus.Errorln(err.Error())
		}
	},
}
