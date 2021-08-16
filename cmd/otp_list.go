package main

import (
	"github.com/fatih/color"
	"github.com/jtaylorcpp/sts"
	"github.com/rodaine/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all issues and associated accounts in STS",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debugln("listing all issuers and accounts in STS")
		entries, err := sts.GetTOTPEntries(stsTableName, stsIssuer, stsAccountName)
		if err != nil {
			logrus.Errorln(err.Error())
		}

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("Issuer", "Account Name", "Roles", "Secondary Authorizer")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, entry := range entries {
			tbl.AddRow(entry.Issuer, entry.AccountName, entry.Roles, entry.SecondaryAuthorization)
		}

		tbl.Print()
	},
}
