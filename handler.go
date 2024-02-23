package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	logger "github.com/Financial-Times/go-logger/v2"
	tidutils "github.com/Financial-Times/transactionid-utils-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Financial-Times/draft-content-suggestions/draft"
	"github.com/Financial-Times/draft-content-suggestions/suggestions"
)

const (
	contentTypeHeader = "Content-Type"
)

type BaseContent struct {
	UUID string `json:"uuid,omitempty"`
}

type requestHandler struct {
	dca draft.ContentAPI
	sua suggestions.UmbrellaAPI
	log *logger.UPPLogger
}

func (rh *requestHandler) draftContentSuggestionsRequest(writer http.ResponseWriter, request *http.Request) {
	uuid := mux.Vars(request)["uuid"]
	log := rh.log.WithTransactionID(tidutils.GetTransactionIDFromRequest(request)).WithUUID(uuid)

	err := ValidateUUID(uuid)
	if err != nil {
		msg := "Invalid UUID"
		log.WithError(err).Warn(msg)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, msg)
		return
	}

	ctx := NewContextFromRequest(request)
	content, err := rh.dca.FetchDraftContent(ctx, uuid)
	if err == draft.ErrDraftNotMappable {
		msg := "Could not provide suggestions for content, as we are unable to map it"
		log.WithError(err).Info(msg)
		_ = WriteJSONMessage(writer, http.StatusUnprocessableEntity, msg)
		return
	}
	if err != nil {
		msg := "Draft content api retrieval has failed."
		log.WithError(err).Error(msg)
		_ = WriteJSONMessage(writer, http.StatusInternalServerError, msg)
		return
	}
	if content == nil {
		msg := "No draft content for UUID"
		log.Warn(msg)
		_ = WriteJSONMessage(writer, http.StatusNotFound, msg)
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)
	if err != nil {
		msg := "Suggestions umbrella api access has failed"
		log.WithError(err).Error(msg)
		_ = WriteJSONMessage(writer, http.StatusServiceUnavailable, msg)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)
	if err != nil {
		// could be related to intermittent/temporary network issues
		// or original Tagme request is no more waiting for a response.
		log.WithError(err).Error("Failed responding to draft content suggestions request")
	}
}

func (rh *requestHandler) getDraftSuggestionsForContent(writer http.ResponseWriter, request *http.Request) {
	log := rh.log.WithTransactionID(tidutils.GetTransactionIDFromRequest(request))

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		msg := "error while reading request body"
		log.WithError(err).Warn(err)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, msg)
		return
	}

	if len(requestBody) == 0 {
		msg := "content body is missing from the request"
		log.Error(msg)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, msg)
		return
	}

	var baseContent BaseContent
	err = json.Unmarshal(requestBody, &baseContent)
	if err != nil {
		msg := "error while unmarshalling uuid from the request payload"
		log.Error(msg)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, msg)
	}

	err = ValidateUUID(baseContent.UUID)
	if err != nil {
		msg := "Invalid payload UUID"
		log.WithError(err).Warn(msg)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, msg)
		return
	}
	log = log.WithUUID(baseContent.UUID)

	contentType := request.Header.Get(contentTypeHeader)
	ctx := NewContextFromRequest(request)

	content, err := rh.dca.FetchValidatedContent(ctx, bytes.NewReader(requestBody), baseContent.UUID, contentType, rh.log)
	if err != nil {
		msg := "failed while validating content"
		log.WithError(err).Warn(msg)
		_ = WriteJSONMessage(writer, http.StatusBadRequest, fmt.Sprintf("%s: %s", msg, err.Error()))
		return
	}

	suggestion, err := rh.sua.FetchSuggestions(ctx, content)
	if err != nil {
		msg := "Suggestions umbrella api access has failed"
		log.WithError(err).Error(msg)
		_ = WriteJSONMessage(writer, http.StatusServiceUnavailable, msg)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(suggestion)
	if err != nil {
		// could be related to intermittent/temporary network issues
		// or original Tagme request is no more waiting for a response.
		log.WithError(err).Error("Failed responding to draft content suggestions request")
	}
}

type message struct {
	Message string `json:"message"`
}

// WriteJSONMessage writes the msg provided as encoded json with the proper content type header added.
func WriteJSONMessage(w http.ResponseWriter, status int, msg string) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	return enc.Encode(&message{Message: msg})
}

// NewContextFromRequest provides a new context including a trxId
// from the request or if missing, a brand new trxId.
func NewContextFromRequest(r *http.Request) context.Context {
	return tidutils.TransactionAwareContext(r.Context(), tidutils.GetTransactionIDFromRequest(r))
}

// ValidateUUID checks the uuid string for supported formats
func ValidateUUID(u string) error {
	_, err := uuid.Parse(u)
	return err
}
