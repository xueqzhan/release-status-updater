package release_status_updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

const (
	DataSetEnv           = "DATASET_ID"
	ProjectIdEnv         = "PROJECT_ID"
	TableNameEnv         = "TABLE_NAME"
	ProductAPIURL        = "https://access.redhat.com/product-life-cycles/api/v1/products"
	OpenshiftProductName = "OpenShift Container Platform 4"
	BQCredentialsFileEnv = "BQ_CREDENTIALS_FILE" // local testing only
	PubSubTopic          = "UpdateReleaseStatusPubSub"
)

func UpdateReleaseStatus(ctx context.Context) error {
	var err error
	projectID := os.Getenv(ProjectIdEnv)
	if len(projectID) == 0 {
		return fmt.Errorf("missing ENV Variable: %s", ProjectIdEnv)
	}

	datasetID := os.Getenv(DataSetEnv)
	if len(datasetID) == 0 {
		return fmt.Errorf("missing ENV Variable: %s", DataSetEnv)
	}

	tableName := os.Getenv(TableNameEnv)
	if len(tableName) == 0 {
		return fmt.Errorf("missing ENV Variable: %s", TableNameEnv)
	}

	credentialsPath := os.Getenv(BQCredentialsFileEnv)
	var client *bigquery.Client
	if len(credentialsPath) > 0 {
		client, err = bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsPath))
	} else {
		client, err = bigquery.NewClient(ctx, projectID)
	}

	if err != nil {
		return fmt.Errorf("error initializing BigQuery client: %v", err)
	}

	// Fetch data from the API
	apiURL := ProductAPIURL
	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("error fetching data from API: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			Name     string `json:"name"`
			Versions []struct {
				Version string `json:"name"`
				Status  string `json:"type"`
			} `json:"versions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("error decoding API response: %v", err)
	}

	for _, product := range data.Data {
		if product.Name == OpenshiftProductName {
			for _, version := range product.Versions {
				queryString := fmt.Sprintf("UPDATE %s.%s.%s SET ReleaseStatus = '%s' WHERE Release = '%s'", projectID, datasetID, tableName, version.Status, version.Version)
				job, err := client.Query(queryString).Run(ctx)
				if err != nil {
					return fmt.Errorf("big query client.Query.Run: %v", err)
				}
				status, err := job.Wait(ctx)
				if err != nil {
					return fmt.Errorf("big query job.Wait: %v", err)
				}
				if status.Err() != nil {
					return fmt.Errorf("big query execution error: %v", status.Err())
				}
			}
		}
	}

	logrus.Infof("BigQuery table updated successfully!")
	return nil
}

func init() {
	functions.CloudEvent(PubSubTopic, updateReleaseStatusPubSub)
}

// MessagePublishedData contains the full Pub/Sub message
type MessagePublishedData struct {
	Message PubSubMessage
}

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// updateReleaseStatusPubSub consumes a CloudEvent message and trigger the UpdateReleaseStatus function.
func updateReleaseStatusPubSub(ctx context.Context, e event.Event) error {
	logrus.Printf("start updating release status")
	var msg MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		logrus.WithError(err).Error("error parsing pub/sub message")
		return fmt.Errorf("event.DataAs: %w", err)
	}

	if err := UpdateReleaseStatus(ctx); err != nil {
		logrus.WithError(err).Error("error updating release status")
		return err
	}
	return nil
}

/*
func main()  {
    UpdateReleaseStatus(context.Background()
}*/
