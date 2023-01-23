// Package api contains code for serving the various API endpoints of the app.
// The code is divided into sub-packages that each correspond to a single API
// endpoint.
package api

import "net/http"

// MethodHandler describes a type that can be used to serve a certain part of an
// API route that corresponds to a specific HTTP method(s). The sub argument is
// the subject of an authentication token (i.e. username) that is validated
// prior.
type MethodHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, sub string)
}
