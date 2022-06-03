package cfn

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

func FetchStackTemplate(ctx context.Context, name string) (io.Reader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("preparing config: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)

	result, err := client.GetTemplate(ctx, &cloudformation.GetTemplateInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching: %w", err)
	}

	return strings.NewReader(*result.TemplateBody), nil
}

func UpdateStackTemplate(ctx context.Context, name string, template io.Reader) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("preparing config: %w", err)
	}

	client := cloudformation.NewFromConfig(cfg)

	rawTemplate, err := io.ReadAll(template)
	if err != nil {
		return fmt.Errorf("buffering: %w", err)
	}

	_, err = client.UpdateStack(ctx, &cloudformation.UpdateStackInput{
		StackName:    aws.String(name),
		TemplateBody: aws.String(string(rawTemplate)),
	})
	if err != nil {
		return fmt.Errorf("updating stack: %w", err)
	}

	waiter := cloudformation.NewStackUpdateCompleteWaiter(client)

	err = waiter.Wait(
		ctx,
		&cloudformation.DescribeStacksInput{StackName: aws.String(name)},
		time.Minute*defaultStackTimeoutMinutes,
	)
	if err != nil {
		return fmt.Errorf("waiting for stack: %w", err)
	}

	return nil
}

const defaultStackTimeoutMinutes = 5
