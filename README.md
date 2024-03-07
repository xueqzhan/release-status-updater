The cloud function in this repository is used to update release status in the BigQuery
openshift-ci-data-analysis.ci_data.Releases table. It is meant to run on regular basis (e.g. daily).

It fetches products from https://access.redhat.com/product-life-cycles/api/v1/products, parses
versions for "OpenShift Container Platform 4" and updates the Releases table with the release status.

Deployment requires editor permissions in the openshift-ci-data-analysis project. The function operates 
on ci_data dataset and so must be deployed in the openshift-ci-data-analysis project. The service account 
openshift-ci-data-writer@openshift-ci-data-analysis.iam.gserviceaccount.com was created ahead of time.
