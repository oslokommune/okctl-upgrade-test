package patch

import (
	"io"
	"strings"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
)

func TestAddBucketVersioning(t *testing.T) {
	testCases := []struct {
		name         string
		withTemplate io.Reader
	}{
		{
			name:         "Should add expected fields",
			withTemplate: strings.NewReader(testTemplate),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := AddBucketVersioning(tc.withTemplate)
			assert.NoError(t, err)

			rawResult, err := io.ReadAll(result)
			assert.NoError(t, err)

			g := goldie.New(t)
			g.Assert(t, tc.name, rawResult)
		})
	}
}

const testTemplate = `AWSTemplateFormatVersion: 2010-09-09
Outputs:
  BucketARN:
    Export:
      Name:
        Fn::Sub: ${AWS::StackName}-BucketARN
    Value:
      Fn::GetAtt:
      - S3Bucket
      - Arn
  S3Bucket:
    Export:
      Name:
        Fn::Sub: ${AWS::StackName}-S3Bucket
    Value:
      Ref: S3Bucket
Resources:
  S3Bucket:
    Properties:
      AccessControl: BucketOwnerFullControl
      BucketName: okctl-mock-cluster-meta
      PublicAccessBlockConfiguration:
        BlockPublicAcls: true
        BlockPublicPolicy: true
        IgnorePublicAcls: true
        RestrictPublicBuckets: true
    Type: AWS::S3::Bucket`
