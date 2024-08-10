package xd_rsync

type MessagePublishInput struct {
	Message        string
	MessageGroupId string
}

type SNSService interface {
	SendMessage(topicArn string, input *MessagePublishInput) error
	SendMessagesBatch(topicArn string, input *[]MessagePublishInput) (int, []error)
}
