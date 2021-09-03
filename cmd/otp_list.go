package main

import (
	"github.com/fatih/color"
	"github.com/jtaylorcpp/sts"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all issues and associated accounts in STS",
	Run: func(cmd *cobra.Command, args []string) {
		logger := sts.GetLogger()
		logger.Debug().Msg("Listing all issuers and accouts in the auth table")
		entries, err := sts.GetTOTPEntries(
			sts.GetStringFlag("table_name"),
			sts.GetStringFlag("issuer"),
			sts.GetStringFlag("account_name"))
		if err != nil {
			logger.Err(err).Msg("Error getting all TOTP entries")
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
