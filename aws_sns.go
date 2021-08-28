package sts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func PublishAuditEvent(issuer, accountName, role string) error {
	logger := GetLogger()
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Str("role", role).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := sns.New(sess)

	msg := "STSFTW has minted new STS crentials:\n" +
		fmt.Sprintf("Issuer: %s\n", issuer) +
		fmt.Sprintf("Account Name: %s\n", accountName) +
		fmt.Sprintf("AWS Role: %s\n", role)

	_, publishErr := svc.Publish(&sns.PublishInput{
		Message:           aws.String(msg),
		MessageAttributes: map[string]*sns.MessageAttributeValue{},
		Subject:           aws.String("Credentials Minted For Account"),
		TopicArn:          aws.String(GetStringFlag("sns_arn")),
	})

	return publishErr
}
