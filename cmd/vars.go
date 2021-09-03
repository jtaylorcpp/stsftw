package main

import "github.com/jtaylorcpp/sts"

var stsEndpoint string
var stsTableName string
var stsIssuer string
var stsAccountName string
var stsRoles []string
var stsRole string
var stsSecondaryAuthorizers []string
var stsSecondaryAuthorizer string
var logLevel string

func init() {
	// root binds
	sts.Bind("table_name", rootCmd.PersistentFlags().Lookup("table-name"))
	sts.Bind("endpoint", rootCmd.PersistentFlags().Lookup("endpoint"))
	sts.Bind("issuer", rootCmd.PersistentFlags().Lookup("issuer"))
	sts.Bind("account_name", rootCmd.PersistentFlags().Lookup("account-name"))
	sts.Bind("log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// enroll binds
	sts.Bind("roles", enrollCmd.PersistentFlags().Lookup("roles"))
	sts.Bind("secondary_authorizers", enrollCmd.PersistentFlags().Lookup("secondary-authorizers"))

	// update binds
	sts.Bind("roles", rolesCmd.PersistentFlags().Lookup("roles"))
	sts.Bind("secondary_authorizers", authorizersCmd.PersistentFlags().Lookup("secondary-authorizers"))

	// get binds
	sts.Bind("secondary_authorizer", getCmd.PersistentFlags().Lookup("secondary-authorizer"))
	sts.Bind("role", getCmd.PersistentFlags().Lookup("role"))
}
