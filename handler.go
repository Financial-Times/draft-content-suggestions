package main

import (
	"fmt"
	"net/http"

	"github.com/Financial-Times/draft-content-suggestions/commons"
	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
	. "github.com/Financial-Times/transactionid-utils-go"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
}

func (rh *requestHandler) draftContentSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {

	uuid := mux.Vars(request)["uuid"]
	logger := log.WithField(TransactionIDKey, GetTransactionIDFromRequest(request)).WithField("uuid", uuid)

	err := commons.ValidateUUID(uuid)

	if err != nil {
		logger.WithError(err).Warn("Invalid UUID")
		commons.WriteJSONMessage(writer, http.StatusBadRequest, "Invalid UUID")
		return
	}

	ctx := commons.NewContextFromRequest(request)

	content, err := rh.dca.FetchDraftContent(ctx, uuid)

	if err == draft.ErrDraftNotMappable {
		log.WithError(err).Info("Could not provide suggestions for content, as we are unable to map it")
		commons.WriteJSONMessage(writer, http.StatusUnprocessableEntity, "Could not provide suggestions for content, as we are unable to map it")
		return
	}

	if err != nil {
		log.WithError(err).Error("Draft content api retrieval has failed.")
		commons.WriteJSONMessage(writer, http.StatusInternalServerError, "Draft content api retrieval has failed.")
		return
	}

	if content == nil {
		log.Warn("No draft content found, cannot provide suggestions")
		commons.WriteJSONMessage(writer, http.StatusNotFound, fmt.Sprintf("No draft content for uuid: %v", uuid))
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)

	if err != nil {
		log.WithError(err).Error("Suggestions umbrella api access has failed")
		commons.WriteJSONMessage(writer, http.StatusServiceUnavailable, "Suggestions umbrella api access has failed")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)

	// could be related to intermittent/temporary network issues
	// or original Tagme request is no more waiting for a response.
	if err != nil {
		log.WithError(err).Error("Failed responding to draft content suggestions request")
	}

}
