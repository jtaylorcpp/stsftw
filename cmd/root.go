package main

import (
	"github.com/jtaylorcpp/sts"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&stsTableName, "table-name", "", "DynamoDB table name used for user tracking")

	rootCmd.PersistentFlags().StringVar(&stsEndpoint, "endpoint", "", "Endpoint to connect to STS API")

	rootCmd.PersistentFlags().StringVar(&stsIssuer, "issuer", "", "TOTP Issuer")

	rootCmd.PersistentFlags().StringVar(&stsAccountName, "account-name", "", "TOTP Account Name")
}

var rootCmd = &cobra.Command{
	Use:   "sts",
	Short: "STS is Simple TOTP STS",
	Long:  `A more simple way to interact with AWS STS`,
	/*Run: func(cmd *cobra.Command, args []string) {
	  // Do Stuff Here
	},*/
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger := sts.GetLogger()
		logger.Err(err).Msg("Error executing cli")
	}
}

func main() {
	Execute()
}
