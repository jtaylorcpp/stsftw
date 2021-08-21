package sts

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/pquerna/otp"
)

type TOTPEntry struct {
	Issuer                 string   `json:"issuer"`
	AccountName            string   `json:"account_name"`
	URL                    string   `json:"url"`
	Roles                  []string `json:"roles"`
	SecondaryAuthorization []string `json:"secondary_authorization"`
}

func NewTOTPEntry(key *otp.Key) (TOTPEntry, error) {
	return TOTPEntry{
		Issuer:      key.Issuer(),
		AccountName: key.AccountName(),
		URL:         key.URL(),
	}, nil
}

func EntryToTOTP(entry TOTPEntry) (*otp.Key, error) {
	return otp.NewKeyFromURL(entry.URL)
}

func AddTOTPEntryToTable(tableName string, entry TOTPEntry) error {
	logger.Trace().Str("issuer", entry.Issuer).Str("account_name", entry.AccountName).Strs("roles", entry.Roles).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		logger.Error().Err(err).Msg("Error unmarshaling entrty")
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		logger.Error().Err(err).Msg("Error writting item to dynamodb")
		return err
	}

	return nil
}

func GetTOTPEntry(tableName, issuer, accountName string) (TOTPEntry, error) {
	if tableName == "" || issuer == "" || accountName == "" {
		return TOTPEntry{}, errors.New("TableName, Issuer, or Account Name are missing for get operation")
	}

	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Str("table", tableName).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return TOTPEntry{}, sessErr
	}

	svc := dynamodb.New(sess)

	filter := expression.And(
		expression.Name("issuer").Equal(expression.Value(issuer)),
		expression.Name("account_name").Equal(expression.Value(accountName)),
	)

	scanExpression := expression.NewBuilder().WithFilter(filter)
	projection := expression.NamesList(
		expression.Name("issuer"),
		expression.Name("account_name"),
		expression.Name("roles"),
		expression.Name("secondary_authorization"),
		expression.Name("url"),
	)

	expr, expressionErr := scanExpression.WithProjection(projection).Build()
	if expressionErr != nil {
		return TOTPEntry{}, expressionErr
	}

	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	firstEntry := TOTPEntry{}
	scanErr := svc.ScanPages(scanInput,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			for _, result := range page.Items {
				entry := TOTPEntry{}
				marshallErr := dynamodbattribute.UnmarshalMap(result, &entry)
				if marshallErr != nil {
					continue
				}
				firstEntry = entry
			}
			return !lastPage
		})

	if scanErr != nil {
		logger.Error().Err(scanErr).Msg("Error scanning dynamo")
		return TOTPEntry{}, scanErr
	}

	return firstEntry, nil
}

func GetTOTPEntries(tableName, issuer, accountName string) ([]TOTPEntry, error) {
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Str("table", tableName).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return []TOTPEntry{}, sessErr
	}

	svc := dynamodb.New(sess)

	var filter expression.ConditionBuilder = expression.ConditionBuilder{}
	scanExpression := expression.NewBuilder()

	switch issuer == "" {
	case true:
		// issuer is empty
		switch accountName == "" {
		case true:
			// issuer and account name are empty
			// do nothing
		case false:
			// issuer is empty but account name is supplied
			filter = expression.Name("account_name").Equal(expression.Value(accountName))
		}
	case false:
		// issuer is supplied
		switch accountName == "" {
		case true:
			// issuer is supplied and account name is empty
			filter = expression.Name("issuer").Equal(expression.Value(issuer))
		case false:
			// issuer is supplied and account name is supplied
			filter = expression.And(
				expression.Name("issuer").Equal(expression.Value(issuer)),
				expression.Name("account_name").Equal(expression.Value(accountName)),
			)
		}
	}

	if accountName != "" || issuer != "" {
		scanExpression.WithFilter(filter)
	}

	projection := expression.NamesList(
		expression.Name("issuer"),
		expression.Name("account_name"),
		expression.Name("roles"),
		expression.Name("secondary_authorization"),
	)

	expr, err := scanExpression.WithProjection(projection).Build()

	if err != nil {
		logger.Error().Err(err).Msg("Error creating dynamo expression")
		return []TOTPEntry{}, err
	}

	scanInput := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	totpEntries := []TOTPEntry{}
	scanErr := svc.ScanPages(scanInput,
		func(page *dynamodb.ScanOutput, lastPage bool) bool {
			for _, result := range page.Items {
				entry := TOTPEntry{}
				marshallErr := dynamodbattribute.UnmarshalMap(result, &entry)
				if marshallErr != nil {
					continue
				}
				totpEntries = append(totpEntries, entry)
			}
			return !lastPage
		})

	if scanErr != nil {
		logger.Error().Err(err).Msg("Error scanning dynamo")
		return []TOTPEntry{}, scanErr
	}

	return totpEntries, nil
}

func UpdateTOTPEntryRoles(tableName, issuer, accountName string, roles []string, replace bool) error {
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Strs("roles", roles).Str("table", tableName).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := dynamodb.New(sess)

	update := expression.Set(
		expression.Name("roles"),
		expression.Value(roles),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		logger.Error().Err(err).Msg("Error building dynamo expression")
		return err
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"issuer": {
				S: aws.String(issuer),
			},
			"account_name": {
				S: aws.String(accountName),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: expr.Update(),
	}

	_, updateErr := svc.UpdateItem(input)

	if updateErr != nil {
		logger.Error().Err(updateErr).Msg("Error updating dynamo item")
		return updateErr
	}

	return nil
}

func UpdateTOTPEntrySecondaryAuthorizations(tableName, issuer, accountName string, authorizations []string, replace bool) error {
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Str("table", tableName).Strs("authorizers", authorizations).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := dynamodb.New(sess)

	update := expression.Set(
		expression.Name("secondary_authorization"),
		expression.Value(authorizations),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		logger.Error().Err(err).Msg("Error creating dynamo expression")
		return err
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"issuer": {
				S: aws.String(issuer),
			},
			"account_name": {
				S: aws.String(accountName),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: expr.Update(),
	}

	_, updateErr := svc.UpdateItem(input)

	if updateErr != nil {
		logger.Error().Err(updateErr).Msg("Error updating dynamo item")
		return updateErr
	}

	return nil
}

func UpdateTOTPEntryMFADevice(tableName, issuer, accountName string, key *otp.Key) error {
	logger.Trace().Str("issuer", issuer).Str("account_name", accountName).Msg("Getting AWS credentials")

	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String(GetStringFlag("region")),
	})

	if sessErr != nil {
		return sessErr
	}

	svc := dynamodb.New(sess)

	update := expression.Set(
		expression.Name("url"),
		expression.Value(key.URL()),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		logger.Error().Err(err).Msg("Error building dynamo expression")
		return err
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"issuer": {
				S: aws.String(issuer),
			},
			"account_name": {
				S: aws.String(accountName),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: expr.Update(),
	}

	_, updateErr := svc.UpdateItem(input)

	if updateErr != nil {
		logger.Error().Err(updateErr).Msg("Error updating dynamo item")
		return updateErr
	}

	return nil
}
