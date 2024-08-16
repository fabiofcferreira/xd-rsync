package main

import (
	"fmt"
	"time"

	xd_rsync "github.com/fabiofcferreira/xd-rsync"
	"github.com/fabiofcferreira/xd-rsync/aws/sns"
	"github.com/fabiofcferreira/xd-rsync/database"
	"github.com/fabiofcferreira/xd-rsync/logger"
	"github.com/fabiofcferreira/xd-rsync/tickers"
)

func captureProductChanges(app *xd_rsync.XdRsyncInstance) tickers.TickerAction {

	var lastSyncTimestamp time.Time
	lastSyncTimestamp, err := time.Parse("2006-01-02T15:04:05", "1900-01-01T00:00:00")
	if err != nil {
		app.Logger.Fatal("failed_set_initial_sync_timestamp", "Failed to set initial sync timestamp", &map[string]interface{}{
			"error": err,
		})
	}

	return func() []error {
		app.Logger.Info("init_send_changed_products_events", "Starting process to send events for updated products", nil)

		allPricedProducts, err := app.Services.Database.GetPricedProductsSinceTimestamp(&lastSyncTimestamp)
		lastSyncTimestamp = time.Now()
		if err != nil {
			app.Logger.Error("failed_get_priced_products", "Failed to get priced products", &map[string]interface{}{
				"error": err,
			})

			return []error{err}
		}

		updatedProductsEvents := []xd_rsync.MessagePublishInput{}
		updatedProductsSkus := []string{}
		for _, product := range *allPricedProducts {

			productDto, err := product.ToJSON()
			if err != nil {
				app.Logger.Error("failed_get_product_dto", "Failed to get product DTO for SNS topic message", &map[string]interface{}{
					"error":   err,
					"sku":     product.SKU,
					"product": product,
				})

				return []error{err}
			}

			updatedProductsSkus = append(updatedProductsSkus, product.SKU)

			updatedProductsEvents = append(updatedProductsEvents, xd_rsync.MessagePublishInput{
				Message:        productDto,
				MessageGroupId: product.SKU,
			})
		}

		if len(updatedProductsEvents) == 0 {
			app.Logger.Info("skip_send_changed_products_events", "No products were changed since last check", nil)

			return []error{err}
		}

		app.Logger.Info("count_product_change_events", "Got all product change events", &map[string]interface{}{
			"changedProductsCount": len(updatedProductsEvents),
			"skus":                 updatedProductsSkus,
		})
		successfulMessages, errors := app.Services.SNS.SendMessagesBatch(app.Config.Queues.ProductUpdatesSnsQueueArn, &updatedProductsEvents)
		if len(errors) > 0 {
			app.Logger.Info("failed_changed_product_events", "Failed to publish updated product event", &map[string]interface{}{
				"error": errors,
			})

			return errors
		}

		app.Logger.Info("finished_changed_product_events", "Finished sending changed products' events", &map[string]interface{}{
			"changedProductsCount":    len(updatedProductsEvents),
			"successfulMessagesCount": successfulMessages,
		})

		return nil
	}
}

func main() {
	cfg, err := GetConfig()
	if err != nil {
		panic(err)
	}

	logger, err := logger.CreateLogger(
		&logger.LoggerOptions{
			IsProduction:      cfg.IsProductionMode,
			InitialFields:     *cfg.DatadogConfig.EventBaseFields,
			DatadogIngestHost: cfg.DatadogConfig.IngestHost,
			DatadogApiKey:     cfg.DatadogConfig.ApiKey,
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
		app.Logger.Fatal("failed_to_create__client", "Failed to create  client", &map[string]interface{}{
			"error": err,
		})
	} else {
		app.Services.Database = dbService
	}

	snsClient, err := sns.CreateClient(&sns.SNSClientCreationInput{
		Region: cfg.AwsRegion,
		Logger: logger,
	})
	if err != nil {
		app.Logger.Fatal("failed_to_create__client", "Failed to create  client", &map[string]interface{}{
			"error": err,
		})
	} else {
		app.Services.SNS = snsClient
	}

	app.Logger.Info("startup_complete", "XD Rsync startup completed", nil)

	tickers.RunEvery(app.Config.SyncFrequency, captureProductChanges(app))
	app.Logger.Info("init_shutdown", "XD Rsync shutdown", nil)
}
