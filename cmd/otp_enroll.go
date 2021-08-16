package main

import (
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
		logrus.Infof("Enrolling new device in STS with issuer <%s> and account name <%s>\n", sts.GetStringFlag("issuer"), sts.GetStringFlag("account_name"))
		key, err := sts.GenerateNewTOTP(sts.GetStringFlag("issuer"), sts.GetStringFlag("account_name"))
		if err != nil {
			logrus.Errorln(err.Error())
		}
		logrus.Infoln(key.String())

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

		logrus.Infoln("TOTP enrolled and verified")

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
