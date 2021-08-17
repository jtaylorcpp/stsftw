package sts

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func PublishAuditEvent(issuer, accountName, role string) error {
	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := sns.New(sess)

	_, publishErr := svc.Publish(&sns.PublishInput{
		Message:           aws.String(""),
		MessageAttributes: map[string]*sns.MessageAttributeValue{},
		Subject:           aws.String(""),
		TopicArn:          aws.String(""),
	})

	return publishErr
}
