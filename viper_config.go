package sts

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config *viper.Viper

func init() {
	config = viper.New()

	config.SetEnvPrefix("sts")

	config.BindEnv("table_name")
	config.BindEnv("endpoint")
	config.BindEnv("issuer")
	config.BindEnv("account_name")
	config.BindEnv("role")
	config.BindEnv("secondary_account_name")
	config.BindEnv("roles")
	config.BindEnv("secondary_authorizers")
	config.BindEnv("sns_arn")
	config.BindEnv("region")
	config.BindEnv("log_level")
	config.SetDefault("log_level", "warn")
}

func Bind(configName string, flag *pflag.Flag) {
	config.BindPFlag(configName, flag)
}

func GetStringFlag(flag string) string {
	return config.GetString(flag)
}

func GetStringArrayFlag(flag string) []string {
	return config.GetStringSlice(flag)
}
