package sns

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/fabiofcferreira/xd-rsync/logger"
)

type SNSClient struct {
	client *sns.Client
	logger *logger.Logger
}

type SNSClientCreationInput struct {
	Region string
	Logger *logger.Logger
}

type MessagePublishSuccess struct {
	Id               string
	MessagePublishId string
}

type MessagePublishError struct {
	eventMessage string
	message      string
}

func (me MessagePublishError) Error() string {
	return fmt.Sprintf("could not publish event to SNS topic. event message: %s error message: %s", me.eventMessage, me.message)
}

func CreateClient(input *SNSClientCreationInput) (*SNSClient, error) {
	clientInstance := &SNSClient{
		logger: input.Logger,
	}

	clientInstance.logger.Info("init_sns_client_create", "Creating SNS client instance", nil)
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		clientInstance.logger.Info("failed_sns_client_create", "Failed to create SNS client instance", &map[string]interface{}{
			"error": err,
		})
		return nil, err
	}

	clientInstance.client = sns.NewFromConfig(sdkConfig)

	clientInstance.logger.Info("finished_sns_client_create", "Created SNS client instance", nil)
	return clientInstance, nil
}

func (s SNSClient) publishMessage(topicArn string, input *xd_rsync.MessagePublishInput, maxRetries int) (*MessagePublishSuccess, error) {
	publishInput := &sns.PublishInput{
		TopicArn:       &topicArn,
		Message:        aws.String(input.Message),
		MessageGroupId: aws.String(input.MessageGroupId),
	}

	var err error
	var publishOutput *sns.PublishOutput
	for tries := 1; tries < maxRetries; tries++ {
		publishOutput, err = s.client.Publish(context.TODO(), publishInput)

		if err == nil && len(*publishOutput.MessageId) > 0 {
			return &MessagePublishSuccess{
				Id:               "msg",
				MessagePublishId: *publishOutput.MessageId,
			}, nil
		}
	}

	return nil, &MessagePublishError{
		eventMessage: input.Message,
		message:      err.Error(),
	}
}

func (s SNSClient) publishMessageList(topicArn string, messages *[]xd_rsync.MessagePublishInput, maxRetries int) (int, []error) {
	allMessageIds := []string{
		"http-request",
	}
	allErrors := []error{}
	errorsMapByMessageId := map[string]error{}

	pendingMessages := []types.PublishBatchRequestEntry{}
	for index, msg := range *messages {
		pendingMessages = append(pendingMessages, types.PublishBatchRequestEntry{
			Id:             aws.String("msg-" + strconv.Itoa(index)),
			Message:        aws.String(msg.Message),
			MessageGroupId: aws.String(msg.MessageGroupId),
		})
		allMessageIds = append(allMessageIds, "msg-"+strconv.Itoa(index))
	}

	batchPublishOutput := &sns.PublishBatchOutput{}
	var batchRequestErr error
	for tries := 0; tries < maxRetries; tries++ {
		publishInput := &sns.PublishBatchInput{
			TopicArn:                   &topicArn,
			PublishBatchRequestEntries: pendingMessages,
		}

		batchPublishOutput, batchRequestErr = s.client.PublishBatch(context.TODO(), publishInput)
		if batchRequestErr != nil {
			continue
		}

		if len(batchPublishOutput.Failed) == 0 {
			return len(*messages), nil
		}

		// Remove published messages from pending list
		for _, publishedMsg := range batchPublishOutput.Successful {
			for pendingMsgIndex, pendingMsg := range pendingMessages {
				if pendingMsg.Id == publishedMsg.Id {
					pendingMessages = append(pendingMessages[:pendingMsgIndex], pendingMessages[pendingMsgIndex+1:]...)
					errorsMapByMessageId[*pendingMsg.Id] = nil
					break
				}
			}
		}

		// Add failed message error description to the map
		for _, failedMsg := range batchPublishOutput.Failed {
			for _, pendingMsg := range pendingMessages {
				if pendingMsg.Id == failedMsg.Id {
					errorsMapByMessageId[*pendingMsg.Id] = errors.New(*failedMsg.Message)
					break
				}
			}
		}
	}

	for _, messageId := range allMessageIds {
		allErrors = append(allErrors, errorsMapByMessageId[messageId], batchRequestErr)
	}
	return len(*messages), allErrors
}

func (s SNSClient) chunkMessages(messages *[]xd_rsync.MessagePublishInput) *map[int][]xd_rsync.MessagePublishInput {
	chunks := &map[int][]xd_rsync.MessagePublishInput{}

	chunkIndex := 0
	for _, message := range *messages {
		if _, ok := ((*chunks)[chunkIndex]); !ok {
			(*chunks)[chunkIndex] = []xd_rsync.MessagePublishInput{}
		}

		(*chunks)[chunkIndex] = append((*chunks)[chunkIndex], message)
		if len((*chunks)[chunkIndex]) == 10 {
			chunkIndex++
		}
	}

	return chunks
}

func (s SNSClient) SendMessage(topicArn string, input *xd_rsync.MessagePublishInput) error {
	s.logger.Info("init_sns_message_send", "Start sending SNS message", &map[string]interface{}{
		"message": input,
	})
	MessagePublishSuccess, err := s.publishMessage(topicArn, input, 5)
	if err != nil {
		s.logger.Error("failed_sns_message_send", "Failed to send SNS message", &map[string]interface{}{
			"error": err,
		})

		return err
	}

	s.logger.Info("finished_sns_message_send", "Finished sending SNS message", &map[string]interface{}{
		"message":          input,
		"messagePublishId": MessagePublishSuccess.MessagePublishId,
	})
	return nil
}

func (s SNSClient) SendMessagesBatch(topicArn string, messages *[]xd_rsync.MessagePublishInput) (int, []error) {
	allErrors := []error{}
	chunks := s.chunkMessages(messages)

	s.logger.Info("init_sns_messages_batch_send", "Start sending batch of SNS messages", &map[string]interface{}{
		"messagesCount": len(*messages),
		"chunks":        len(*chunks),
	})

	wg := sync.WaitGroup{}

	totalSentMessages := 0
	for chunkNumber := 0; chunkNumber < len(*chunks); chunkNumber++ {
		wg.Add(1)

		go func() {
			currentChunk := (*chunks)[chunkNumber]
			sentMessages, errs := s.publishMessageList(topicArn, &currentChunk, 5)
			if len(errs) > 0 {
				s.logger.Warn("sns_messages_batch_with_errors", "SNS messages batch publish with errors", &map[string]interface{}{
					"errors": errs,
				})
				allErrors = append(allErrors, errs...)
			}

			totalSentMessages += sentMessages
			wg.Done()
		}()
	}

	wg.Wait()

	s.logger.Info("finished_sns_messages_batch_send", "Finished sending batch of SNS messages", &map[string]interface{}{
		"messagesCount": len(*messages),
		"chunks":        len(*chunks),
		"errors":        allErrors,
	})

	return totalSentMessages, allErrors
}
