package main

import (
	"context"
	"fmt"
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

	responseCh := make(chan []byte)
	errorCh := make(chan error)

	now := time.Now()
	ctx, cancelCtx := context.WithTimeout(commons.NewContextFromRequest(request), now.Add(rh.srt).Sub(now))
	defer cancelCtx()

	go func(ctx context.Context, responseCh chan []byte, errorCh chan error) {
		content, err := rh.dca.FetchDraftContent(ctx, uuid)
		if err != nil {
			errorCh <- err
			return
		}
		responseCh <- content
	}(ctx, responseCh, errorCh)

	var content []byte

	select {
	case <-ctx.Done():
		log.WithField("uuid", uuid).Error("Timed out processing draft content suggestions request during fetching draft content")
		commons.WriteJSONMessage(writer, http.StatusRequestTimeout, "Draft content api access has timed out.")
		return
	case err := <-errorCh:
		log.WithError(err).WithField("uuid", uuid).Error("Draft content api access has failed.")
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, "Draft content api access has failed.")
		return
	case content = <-responseCh:
		if content == nil {
			commons.WriteJSONMessage(writer, http.StatusNotFound, fmt.Sprintf("No draft content for uuid: %v", uuid))
			return
		}
	}

	go func(ctx context.Context, responseChan chan []byte, errorCh chan error) {
		suggestion, err := rh.sua.FetchSuggestions(ctx, content)
		if err != nil {
			errorCh <- err
			return
		}
		responseCh <- suggestion
	}(ctx, responseCh, errorCh)

	select {
	case <-ctx.Done():
		log.WithField("uuid", uuid).Error("Timed out processing draft content suggestions request during suggestions umbrella api access")
		commons.WriteJSONMessage(writer, http.StatusRequestTimeout, "Suggestions Umbrella api access has timed out.")
		return
	case err := <-errorCh:
		log.WithError(err).WithField("uuid", uuid).Error("Suggestions umbrella api access has failed")
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, "Suggestions umbrella api access has failed")
		return
	case suggestion := <-responseCh:
		writer.Header().Set("Content-Type", "application/json")
		_, err = writer.Write(suggestion)

		if err != nil {
			log.WithError(err).WithField("uuid", uuid).Error("Failed responding to draft content suggestions request")
		}
	}
}
