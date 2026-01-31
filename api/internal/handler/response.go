package handler

import (
	"encoding/json"
	"net/http"

	"github.com/forgo/saga/api/internal/model"
)

// DataResponse wraps a successful response with optional HATEOAS links
type DataResponse struct {
	Data  interface{}       `json:"data"`
	Links map[string]string `json:"_links,omitempty"`
}

// CollectionResponse wraps a collection response with pagination
type CollectionResponse struct {
	Data       interface{}       `json:"data"`
	Pagination *PaginationInfo   `json:"pagination,omitempty"`
	Links      map[string]string `json:"_links,omitempty"`
}

// PaginationInfo contains cursor-based pagination info
type PaginationInfo struct {
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// WriteData writes a successful data response
func WriteData(w http.ResponseWriter, status int, data interface{}, links map[string]string) {
	response := DataResponse{
		Data:  data,
		Links: links,
	}
	WriteJSON(w, status, response)
}

// WriteCollection writes a collection response with pagination
func WriteCollection(w http.ResponseWriter, status int, data interface{}, pagination *PaginationInfo, links map[string]string) {
	response := CollectionResponse{
		Data:       data,
		Pagination: pagination,
		Links:      links,
	}
	WriteJSON(w, status, response)
}

// WriteError writes an error response using RFC 9457 Problem Details
func WriteError(w http.ResponseWriter, err *model.ProblemDetails) {
	WriteJSON(w, err.Status, err)
}

// DecodeJSON decodes a JSON request body into the given struct
func DecodeJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
