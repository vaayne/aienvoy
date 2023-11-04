package llm

import (
	"context"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func newClaude() *claude.Claude {
	return claude.New(getAWSConfig(), newDao())
}

func getAWSConfig() aws.Config {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(config.GetConfig().AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.GetConfig().AWS.AccessKeyId,
			config.GetConfig().AWS.SecretAccessKey,
			"",
		)))
	if err != nil {
		slog.Error("get aws config error", "err", err)
		return aws.Config{}
	}
	return cfg
}
