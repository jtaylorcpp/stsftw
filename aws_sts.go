package sts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func GetCredentials(issuer, accountName, role string) (STSCredentials, error) {
	logger := GetLogger()
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Str("role", role).Msg("Getting AWS credentials")
	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return STSCredentials{}, sessErr
	}

	svc := sts.New(sess)

	callerIdenity, identityErr := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if identityErr != nil {
		logger.Error().Err(identityErr).Msg("Error calling GetCallerIdentity")
		return STSCredentials{}, identityErr
	}

	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", *callerIdenity.Account, role)),
		RoleSessionName: aws.String(fmt.Sprintf("%s.%s", issuer, accountName)),
	}

	creds, assumeErr := svc.AssumeRole(input)
	if assumeErr != nil {
		logger.Error().Err(assumeErr).Msg("Error assuming role")
		return STSCredentials{}, assumeErr
	}

	return STSCredentials{
		AccessKeyId:     *creds.Credentials.AccessKeyId,
		SecretAccessKey: *creds.Credentials.SecretAccessKey,
		SessionToken:    *creds.Credentials.SessionToken,
	}, nil
}

type STSCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
}
