package main

import (
	"fmt"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/fabiofcferreira/xd-rsync/aws/sns"
	"github.com/fabiofcferreira/xd-rsync/database"
	"github.com/fabiofcferreira/xd-rsync/logger"
)

func rsyncInDaemonMode(app *xd_rsync.XdRsyncInstance) {
	ticker := time.NewTicker(app.Config.SyncFrequency)
	done := make(chan bool)

	var lastCheckTimestamp *time.Time = nil
	go func() {
		for {
			<-ticker.C

			app.Logger.Info("init_send_changed_products_events", "Starting process to send events for updated products", &map[string]interface{}{
				"minimumUpdatedAt": lastCheckTimestamp,
			})

			products, err := app.Services.Database.GetPricedProductsSinceTimestamp(lastCheckTimestamp)
			now := time.Now()
			lastCheckTimestamp = &now
			if err != nil {
				app.Logger.Error("failed_get_priced_products", "Failed to get priced products", &map[string]interface{}{
					"error": err,
				})
				continue
			}

			updatedProductsEvents := []string{}
			for _, product := range *products {
				productDto, err := product.ToJSON()
				if err != nil {
					app.Logger.Error("failed_get_product_dto", "Failed to get product DTO for SNS topic message", &map[string]interface{}{
						"error":   err,
						"product": product,
					})
					continue
				}

				updatedProductsEvents = append(updatedProductsEvents, productDto)
			}

			app.Logger.Info("count_changed_products_events", "Got all changed product events", &map[string]interface{}{
				"changedProductsCount": len(updatedProductsEvents),
			})

			if len(updatedProductsEvents) == 0 {
				app.Logger.Info("skip_send_changed_products_events", "No products were changed since last check", nil)
				continue
			}

			successfulMessages, errors := app.Services.SNS.SendMessagesBatch(app.Config.Queues.ProductUpdatesSnsQueueArn, updatedProductsEvents)
			if len(errors) > 0 {
				app.Logger.Info("failed_changed_product_events", "Failed to publish updated product event", &map[string]interface{}{
					"error": errors,
				})
			}

			app.Logger.Info("finished_changed_product_events", "Finished sending changed products' events", &map[string]interface{}{
				"changedProductsCount":    len(updatedProductsEvents),
				"successfulMessagesCount": successfulMessages,
			})
		}
	}()

	done <- true
}

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logger.CreateLogger(
		&logger.LoggerOptions{
			IsProduction:  cfg.IsProductionMode,
			InitialFields: *cfg.DatadogConfig.EventBaseFields,
			DatadogApiKey: cfg.DatadogConfig.DatadogApiKey,
		})
	if err != nil {
		panic(fmt.Errorf("logger error: %w", err))
	}

	app := &xd_rsync.XdRsyncInstance{
		Config:   cfg,
		Logger:   logger,
		Services: &xd_rsync.XdRsyncServices{},
	}

	dbService, err := database.CreateClient(&database.DatabaseClientCreationInput{
		DSN:    cfg.DSN,
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}

	app.Services.Database = dbService

	snsClient, err := sns.CreateClient(&sns.SNSClientCreationInput{
		Region: cfg.AwsRegion,
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}

	app.Services.SNS = snsClient

	app.Logger.Info("startup_complete", "XD Rsync startup completed", nil)

	rsyncInDaemonMode(app)
	app.Logger.Info("init_shutdown", "XD Rsync shutdown", nil)
}
