package policies

import (
	_ "embed" //nolint:revive
	"fmt"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/cfn"
)

const (
	defaultLogicalBucketPolicyName   = "LokiS3ServiceAccountPolicy"
	defaultBucketPolicyOutputName    = "LokiS3ServiceAccountPolicy"
	defaultLogicalDynamoDBPolicyName = "LokiDynamoDBServiceAccountPolicy"
	defaultDynamoDBPolicyOutputName  = "LokiDynamoDBServiceAccountPolicy"
)

// CreateS3BucketPolicy knows how to create a S3 bucket policy allowing Loki to do necessary operations
func CreateS3BucketPolicy(ctx context.Context, clusterName string, bucketARN string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx.Ctx)
	if err != nil {
		return "", fmt.Errorf("preparing config: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)
	stackName := fmt.Sprintf("okctl-s3bucketpolicy-%s-loki", clusterName)

	if ctx.Flags.DryRun {
		return "arn:to:be:calculated:for:bucketpolicy", nil
	}

	err = createBucketPolicyStack(ctx.Ctx, client, clusterName, stackName, bucketARN)
	if err != nil {
		return "", fmt.Errorf("creating stack: %w", err)
	}

	out, err := client.DescribeStacks(ctx.Ctx, &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)})
	if err != nil {
		return "", fmt.Errorf("describing stack: %w", err)
	}

	arn, err := cfn.GetOutput(out, defaultLogicalBucketPolicyName, defaultBucketPolicyOutputName)
	if err != nil {
		return "", fmt.Errorf("getting ARN: %w", err)
	}

	return arn, nil
}

// CreateDynamoDBPolicy knows how to create a DynamoDB policy allowing Loki to do necessary operations
func CreateDynamoDBPolicy(ctx context.Context, awsAccountID string, awsRegion string, clusterName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx.Ctx)
	if err != nil {
		return "", fmt.Errorf("preparing config: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)
	stackName := fmt.Sprintf("okctl-dynamodbpolicy-%s-loki", clusterName)

	if ctx.Flags.DryRun {
		return "arn:to:be:calculated:for:dynamodbpolicy", nil
	}

	err = createDynamoDBPolicyStack(createDynamoDBPolicyStackOpts{
		ctx:          ctx.Ctx,
		client:       client,
		stackName:    stackName,
		awsAccountID: awsAccountID,
		awsRegion:    awsRegion,
		clusterName:  clusterName,
	})
	if err != nil {
		return "", fmt.Errorf("creating stack: %w", err)
	}

	out, err := client.DescribeStacks(ctx.Ctx, &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)})
	if err != nil {
		return "", fmt.Errorf("describing stack: %w", err)
	}

	arn, err := cfn.GetOutput(out, defaultLogicalDynamoDBPolicyName, defaultDynamoDBPolicyOutputName)
	if err != nil {
		return "", fmt.Errorf("getting ARN: %w", err)
	}

	return arn, nil
}
