# draft-content-suggestions

[![Circle CI](https://circleci.com/gh/Financial-Times/draft-content-suggestions/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/draft-content-suggestions/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/draft-content-suggestions)](https://goreportcard.com/report/github.com/Financial-Times/draft-content-suggestions) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/draft-content-suggestions/badge.svg)](https://coveralls.io/github/Financial-Times/draft-content-suggestions)

## Introduction
Draft Content Suggestions as a microservice, provides consolidated suggestions via fetching draft content 
from Draft Content service and querying Suggestions Umbrella service.  

## Installation

Download the source code, dependencies and test dependencies:

        go get github.com:Financial-Times/draft-content-suggestions
        cd $GOPATH/src/github.com/Financial-Times/draft-content-suggestions
        go build .

## Running locally

1. Run the tests and install the binary:

        go test ./...
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/draft-content-suggestions [--help]

Options:

        --app-system-code="draft-content-suggestions"            System Code of the application ($APP_SYSTEM_CODE)
        --app-name="Annotation Suggestions API"                   Application name ($APP_NAME)
        --port="8080"                                           Port to listen on ($APP_PORT)
        --draft-content-endpoint="http://localhost:9000/drafts/content" Draft Content Service
        --draft-content-gtg-endpoint="http://localhost:9000/__gtg" Draft Content Health Service
        --suggestions-umbrella-endpoint="http://test.api.ft.com/content/suggest" Suggestions Umbrella Service
        --suggestions-api-key="" Suggestions service apiKey
        

3. Test:

    1. Either using curl:

            curl http://localhost:8080/drafts/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/suggestions | json_pp

    1. Or using [httpie](https://github.com/jkbrzt/httpie):

            http GET http://localhost:8080/drafts/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/suggestions

## Build and deployment
_How can I build and deploy it (lots of this will be links out as the steps will be common)_

* Built by Docker Hub on merge to master: [coco/draft-content-suggestions](https://hub.docker.com/r/coco/draft-content-suggestions/)
* CI provided by CircleCI: [draft-content-suggestions](https://circleci.com/gh/Financial-Times/draft-content-suggestions)

## API

For detailed documentation in OpenAPI format, please see [here](./_ft/api.yml).

### Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
* NOTE: `/__build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.

## Change/Rotate sealed secrets

Please reffer to documentation in [pac-global-sealed-secrets-eks](https://github.com/Financial-Times/pac-global-sealed-secrets-eks/blob/master/README.md). Here are explained details how to create new, change existing sealed secrets.
