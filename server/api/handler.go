package api

import (
	"net/http"

	"server/auth"
)

// Handler is a http.Handler that can be used to handle board requests.
type Handler struct {
	authHeaderReader   auth.HeaderReader
	authTokenValidator auth.TokenValidator
	methodHandlers     map[string]MethodHandler
}

// NewHandler creates and returns a new Handler.
func NewHandler(
	authHeaderReader auth.HeaderReader,
	authTokenValidator auth.TokenValidator,
	methodHandlers map[string]MethodHandler,
) Handler {
	return Handler{
		authHeaderReader:   authHeaderReader,
		authTokenValidator: authTokenValidator,
		methodHandlers:     methodHandlers,
	}
}

// ServeHTTP responds to requests made to the board route.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Keep track of allowed methods to return them in
	// "Access-Control-Allow-Methods" header on 405.
	var allowedMethods []string

	// Find the MethodHandler for the HTTP method of the received request.
	for method, methodHandler := range h.methodHandlers {
		allowedMethods = append(allowedMethods, method)

		// If found, authenticate and handle with MethodHandler.
		if r.Method == method {
			// Get auth token from Authorization header, validate it, and get
			// the subject of the token.
			authToken := h.authHeaderReader.Read(
				r.Header.Get(auth.AuthorizationHeader),
			)
			sub := h.authTokenValidator.Validate(authToken)
			if sub == "" {
				w.Header().Set(auth.WWWAuthenticate())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Token sub is used as the username in methodHandler.Handle.
			methodHandler.Handle(w, r, sub)
			return
		}
	}
	// This path of execution means no MethodHandler was found in
	// h.methodHandlers for the HTTP method of the request.
	w.Header().Add(AllowedMethods(allowedMethods))
	w.WriteHeader(http.StatusMethodNotAllowed)
}
