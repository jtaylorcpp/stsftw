STSFTW
===============

## What?
STSFTW is a service that allows you to link your Google Authenticator (or similar) to an internet-facing service ran in AWS for the purpose of generating AWS STS creds.

This will let you get rid of permanent creds on your dev machine and never run the risk of checking in or publishing usable keys (YIKES).

## How is this different?

STSFTW (this project) uses MFA and Multi-Party Auth to take the place of a tradional password/mfa provided by an Idenity platform.

This means that as a user/operater of this project, you can get rid of all of your local, permanent creds for AWS and rely only on TOTP for you auth. And, in the case you feel like you need more tinfoil, you can enroll another device (or a second profile on the same device) to provide either 2 TOTP code or a genuine Multi-Part Auth.

## How it Works

Each entry in the auth table for this project has the following values:

- *issuer* - The issue is a logical grouping of users
- *account_name* - The name of the user to be added to the issuer group
- *url* - The TOTP URL used to validate the client for the (issuer, account_name) pair
- *roles* - A list of AWS Role Names that the (issuer, account_name) can be granted
- *SecondaryAuthorization* - A list of account_names within the same issuer which can act as the Multi-Party Auth provider

These entries are added to the table by the application admin (or any user with AWS IAM access to this DynamoDB table).

Once the authorization table has been updated and a user enrolled, the user can use the `get` command to get AWS STS creds for the supplied role.

These AWS STS creds are provided by a AWS Lambda, AWS ALB setup in which (for the example case in this repo) has an AWS Route53 entry. This allows for clients to be able to access the API via HTTPS and have a privileged AWS Lambda do the AWS STS call to provide the creds.

## Getting Started

### Infrastructure

Description of app.yml

Where to place

How to run terragrunt

Requirements before running

### Application Setup

#### ENV Vars and CLI Args

The STS client can be configured with both flags and ENV vars. Setting the ENV vars can allow a more simplified experience when using the cli.

| ENV var | cli flag | description |
|---|---|---|
| STS_ISSUER | issuer | Issuer is used for logical group managment. Set value is used for the current operation. |
| STS_ACCOUNT_NAME | account-name | Account name is the user in question for the current operation |
| STS_TABLE_NAME | table-name | Name of the DynamoDB table used for the auth table. |
| STS_ENDPOINT | endpoint | URL of the API endpoint. |
| STS_ROLE | role | Name of the AWS Role to get credentials for. |
| STS_ROLES | roles | List of names of AWS Roles to add to a users auth entry. |
| STS_SECONDARY_ACCOUNT_NAME | secondary-authorizer | *account_name* of the secondary entity used for Multi-Party Auth. |
| STS_SECONDARY_AUTHORIZERS | secondary-authorizers | List of *account_name* entities who can provide Multi-Party Auth for a given *account_name* who all exist within the same *issuer*. |
|| totp-code | TOTP code from an enrolled device (primary). |
|| secondary-totp-code | TOTP code from an enrolled device (secondary). |

##### Setting up ~/.profile

The values for issuer, account name, endpoint, role, and secondary authorizer (if needed) make the most sense to add to your future sessions.

The endpoint variable is one that will have to be gotten from the application operator and issue, account name, role, and secondary are values that are made up when a device is enrolled.

#### Enrolling an admin device

When enrolling a device, the user running the commands will need active AWS creds with permission  to write to the DynamoDB table.

```bash
sts enroll --issuer personal-account --account-name admin
```

For the above example, the table name was set by env var and, for the admin role, no AWS Roles were associated and there are no secondary authorizers.

#### Enrolling the first user device

For enrolling the first user/device needs the same permissions as aboce.

```bash
sts enroll --issuer personal-account --account-name chromebook-admin --roles AccountAdmmin --secondary-authorizers admin
```

This will create a new user within the bounds of the same issuer and assign a custom adim role to it while also forcing the admin user to help with multi-party auth.

#### Getting the first AWS STS creds
