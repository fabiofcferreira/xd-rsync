package xd_rsync

import "github.com/fabiofcferreira/xd-rsync/logger"

type QueuesConfig struct {
	ProductUpdatesSnsQueueArn string `json:"productUpdatesSnsQueueArn,omitempty"`
}

type Config struct {
	Environment      string        `json:"environment"`
	IsProductionMode bool          `json:"isProductionMode"`
	AwsRegion        string        `json:"awsRegion"`
	DSN              string        `json:"dsn"`
	Queues           *QueuesConfig `json:"queues"`
	CloseOnFinish    bool          `json:"closeOnFinish"`
}

type XdRsyncServices struct {
	Database DatabaseService
	SNS      SNSService
}

type XdRsyncInstance struct {
	Config   *Config
	Logger   *logger.Logger
	Services *XdRsyncServices
}
