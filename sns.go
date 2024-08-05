package xd_rsync

type SNSService interface {
	SendMessage(topicArn string, message string) error
	SendMessagesBatch(topicArn string, messages []string) (int, []error)
}
