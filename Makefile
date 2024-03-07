build:
	go build .
.PHONY: build

deploy: build
	gcloud functions deploy UpdateReleaseStatus \
        --gen2 \
		--project openshift-ci-data-analysis \
		--runtime go120 \
		--service-account openshift-ci-data-writer@openshift-ci-data-analysis.iam.gserviceaccount.com \
		--source=. \
		--entry-point=UpdateReleaseStatusPubSub \
		--set-env-vars PROJECT_ID=openshift-ci-data-analysis,DATASET_ID=ci_data,TABLE_NAME=Releases \
		--trigger-topic=UpdateReleaseStatusPubSub \
		--allow-unauthenticated
.PHONY: deploy

create-pubsub-topic:
	gcloud pubsub topics create UpdateReleaseStatusPubSub
.PHONY: create-pubsub-topic

# publish-pubsub-message is used to trigger the update function manually
publish-pubsub-message:
	gcloud pubsub topics publish UpdateReleaseStatusPubSub --message hello
.PHONY: publish-pubsub-message

# schedule-release-status-updater creates a cron job that triggers the update function automatically
schedule-release-status-updater:
	gcloud scheduler jobs create pubsub release-status-updater-job --schedule="0 0 * * *" --topic=UpdateReleaseStatusPubSub --message-body "hello" --location=us-east1
.PHONY: schedule-release-status-updater

delete:
	gcloud scheduler jobs delete release-status-updater-job --location=us-east1
	gcloud functions delete UpdateReleaseStatus --project openshift-ci-data-analysis
	gcloud pubsub topics delete UpdateReleaseStatusPubSub
.PHONY: delete
