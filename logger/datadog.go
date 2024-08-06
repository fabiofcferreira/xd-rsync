package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
)

const DATADOG_INGEST_URL = "https://http-intake.logs.datadoghq.eu/api/v2/logs"

type DatadogIngestClient struct {
	ingestUrl string
}

func (c DatadogIngestClient) transformEventPayloadToJson(eventDetails map[string]interface{}) ([]byte, error) {
	bytes, err := json.Marshal(eventDetails)
	if err != nil {
		return nil, fmt.Errorf("could not transform event to JSON format: %s", err)
	}

	return bytes, nil
}

func (c DatadogIngestClient) SendEvent(eventDetails map[string]interface{}) error {
	eventPayload, err := c.transformEventPayloadToJson(eventDetails)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.ingestUrl, mime.TypeByExtension(".json"), bytes.NewBuffer(eventPayload))
	if err != nil {
		return fmt.Errorf("could not post event into datalog: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("event was not ingested but response body could not be parsed: %s", string(bodyBytes))
		}

		return fmt.Errorf("event was not ingested. api response: %s", string(bodyBytes))
	}

	return err
}

func createDatadogIngestClient(host, apiKey string) (*DatadogIngestClient, error) {
	ingestUrl := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/api/v2/logs",
	}

	queryParams := ingestUrl.Query()
	queryParams.Set("dd-api-key", apiKey)
	queryParams.Set("ddsource", "xd-rsync")
	queryParams.Set("service", "xd-rsync")

	ingestUrl.RawQuery = queryParams.Encode()

	return &DatadogIngestClient{
		ingestUrl: ingestUrl.String(),
	}, nil
}
