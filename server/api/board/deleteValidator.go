package board

import (
	"errors"
	"net/url"
)

// POSTValidator describes a type that can be used to validate the URL query
// parameters sent to the board route on DELETE requests.
type DELETEReqValidator interface {
	Validate(url.Values) (string, error)
}

// POSTValidator can be used to validate the URL query parameters sent to the
// board route on DELETE requests.
type DELETEValidator struct{}

// NewDELETEValidator creates and returns a new DELETEValidator.
func NewDELETEValidator() DELETEValidator { return DELETEValidator{} }

// Validate validates the URL query parameters sent to the board route on DELETE
// requests.
func (v DELETEValidator) Validate(qParams url.Values) (string, error) {
	id := qParams.Get("id")
	if id == "" {
		return "", errEmptyBoardID
	}
	return id, nil
}

// ErrEmptyBoardID is the error that the Validate method of DELETEValidator
// returns when the id URL query parameter is empty.
var errEmptyBoardID = errors.New("Board ID cannot be empty.")
