package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/fabiofcferreira/xd-rsync/logger"
)

type SnsClient struct {
	client *sns.Client
	logger *logger.Logger
}

type ClientInitialisationInput struct {
	Region string
	Logger *logger.Logger
}

func (s *SnsClient) Init(input *ClientInitialisationInput) {
	s.logger = input.Logger
	s.client = sns.New(sns.Options{
		Region: input.Region,
	})
}

func (s *SnsClient) SendMessage(topicArn string, message string) (*string, error) {
	publishInput := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  aws.String(message),
	}

	s.logger.Error("init_sns_message_send", "Start sending SNS message", &map[string]interface{}{
		"message": string(message),
	})
	publishOutput, err := s.client.Publish(context.TODO(), publishInput)
	if err != nil {
		s.logger.Error("failed_sns_message_send", "Failed to send SNS message", &map[string]interface{}{
			"error": err,
		})

		return nil, err
	}

	s.logger.Error("finished_sns_message_send", "Finished sending SNS message", &map[string]interface{}{
		"message":   string(message),
		"messageId": *((*publishOutput).MessageId),
	})
	return (*publishOutput).MessageId, nil
}
