// Package policies exposes a simplified API for dealing with policies
package policies

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type createDynamoDBPolicyStackOpts struct {
	ctx          context.Context
	client       *cloudformation.Client
	stackName    string
	awsAccountID string
	awsRegion    string
	clusterName  string
}
