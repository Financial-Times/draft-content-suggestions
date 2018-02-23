package main

import (
	"fmt"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
}

func (rh *requestHandler) draftContentSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {

	uuid := mux.Vars(request)["uuid"]

	err := commons.ValidateUUID(uuid)

	if err != nil {
		log.WithError(err).Warn("Request with invalid uuid:", uuid)
		commons.WriteJSONMessage(writer, http.StatusBadRequest, "Invalid UUID")
		return
	}

	ctx := commons.NewContextFromRequest(request)

	content, err := rh.dca.FetchDraftContent(ctx, uuid)

	if err != nil {
		log.WithError(err).Error("Draft content api access has failed for uuid:", uuid)
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, fmt.Sprintf("Draft content api access has failed for uuid: %v", uuid))
		return
	}

	if content == nil {
		commons.WriteJSONMessage(writer, http.StatusNotFound, fmt.Sprintf("No draft content for uuid: %v", uuid))
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)

	if err != nil {
		log.WithError(err).Error("Suggestions umbrella api access has failed for uuid:", uuid)
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, fmt.Sprintf("Suggestions umbrella api access has failed for uuid: %v", uuid))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)

	// could be related to intermittent/temporary network issues
	// or original Tagme request is no more waiting for a response.
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed responding to draft content suggestions request for uuid: %s", uuid))
	}

}
