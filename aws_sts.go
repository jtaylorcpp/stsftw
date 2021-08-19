package sts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
)

func GetCredentials(issuer, accountName, role string) (STSCredentials, error) {
	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return STSCredentials{}, sessErr
	}

	svc := sts.New(sess)

	callerIdenity, identityErr := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if identityErr != nil {
		logrus.Errorln(identityErr.Error())
		return STSCredentials{}, identityErr
	}

	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", *callerIdenity.Account, role)),
		RoleSessionName: aws.String(fmt.Sprintf("%s.%s", issuer, accountName)),
	}

	creds, assumeErr := svc.AssumeRole(input)
	if assumeErr != nil {
		logrus.Errorln(assumeErr.Error())
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
