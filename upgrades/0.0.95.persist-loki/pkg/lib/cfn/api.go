package cfn

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

// GenerateTags knows how to generate required(by okctl) tags for CloudFormation resources
func GenerateTags(clusterName string) []types.Tag {
	return []types.Tag{
		{
			Key:   aws.String("alpha.okctl.io/cluster-name"),
			Value: aws.String(clusterName),
		},
		{
			Key:   aws.String("alpha.okctl.io/managed"),
			Value: aws.String("true"),
		},
		{
			Key:   aws.String("alpha.okctl.io/okctl-commit"),
			Value: aws.String("unknown"),
		},
		{
			Key:   aws.String("alpha.okctl.io/okctl-version"),
			Value: aws.String("0.0.94"),
		},
	}
}

// GetOutput knows how to retrieve an exported value from a CloudFormation template
func GetOutput(result *cloudformation.DescribeStacksOutput, _ string, outputName string) (string, error) {
	if len(result.Stacks) != 1 {
		return "", errors.New("unexpected amount of stacks")
	}

	for _, output := range result.Stacks[0].Outputs {
		if *output.OutputKey == outputName {
			return *output.OutputValue, nil
		}
	}

	return "", errors.New("output not found")
}
