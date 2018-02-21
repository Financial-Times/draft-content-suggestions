package main

import (
	"fmt"
	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
}

func (rh *requestHandler) annotationSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {

	uuid := mux.Vars(request)["uuid"]

	err := commons.ValidateUUID(uuid)

	if err != nil {
		commons.WriteJSONMessage(writer, http.StatusBadRequest, fmt.Sprintln("Draft content api access error:", err))
		return
	}

	ctx := commons.NewContextFromRequest(request)

	content, err := rh.dca.FetchDraftContent(ctx, uuid)

	if err != nil {
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, fmt.Sprintln("Draft content api access error:", err))
		return
	}

	if content == nil {
		commons.WriteJSONMessage(writer, http.StatusNotFound, fmt.Sprintln("No draft content for uuid:", uuid))
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)

	if err != nil {
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, fmt.Sprintln("Suggestions umbrella api access error:", err))
		return
	}

	_, err = io.Copy(writer, suggestion)

	// could be related to intermittent/temporary network issues
	// or original Tagme request is no more waiting for a response.
	if err != nil {
		log.WithError(err).Error(fmt.Println("Failed responding to annotation suggestions request for uuid:", uuid))
	}

}
