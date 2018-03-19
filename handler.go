package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
	srt time.Duration
}

func (rh *requestHandler) draftContentSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {

	uuid := mux.Vars(request)["uuid"]
	err := commons.ValidateUUID(uuid)

	if err != nil {
		log.WithError(err).WithField("uuid", uuid).Warn("Invalid UUID")
		commons.WriteJSONMessage(writer, http.StatusBadRequest, "Invalid UUID")
		return
	}

	ctx, cancelCtx := context.WithTimeout(commons.NewContextFromRequest(request), rh.srt)
	defer cancelCtx()

	content, err := rh.dca.FetchDraftContent(ctx, uuid)

	if err != nil {
		if isTimeoutError(err) {
			log.WithError(err).WithField("uuid", uuid).Error("Timed out processing draft content suggestions request during fetching draft content")
			commons.WriteJSONMessage(writer, http.StatusGatewayTimeout, "Draft content api access has timed out.")
			return
		}

		log.WithError(err).WithField("uuid", uuid).Error("Draft content api access has failed.")
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, "Draft content api access has failed.")
		return
	}

	if content == nil {
		commons.WriteJSONMessage(writer, http.StatusNotFound, fmt.Sprintf("No draft content for uuid: %v", uuid))
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)

	if err != nil {
		if isTimeoutError(err) {
			log.WithError(err).WithField("uuid", uuid).Error("Timed out processing draft content suggestions request during suggestions umbrella api access")
			commons.WriteJSONMessage(writer, http.StatusGatewayTimeout, "Suggestions Umbrella api access has timed out.")
			return
		}

		log.WithError(err).WithField("uuid", uuid).Error("Suggestions umbrella api access has failed")
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, "Suggestions umbrella api access has failed")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)

	if err != nil {
		log.WithError(err).WithField("uuid", uuid).Error("Failed responding to draft content suggestions request")
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	if netError, assertion := err.(net.Error); assertion {
		return netError.Timeout()
	}

	return false
}
