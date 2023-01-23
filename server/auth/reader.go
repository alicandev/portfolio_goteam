package auth

import "strings"

// BearerTokenReader can be used to read an Authorization header value that
// contains a Bearer token.
type BearerTokenReader struct{}

// NewBearerTokenReader creates and returns a new BearerTokenReader.
func NewBearerTokenReader() BearerTokenReader { return BearerTokenReader{} }

// Read reads an Authorization header value that contains a Bearer token and
// returns the token. If anything goes wrong, an empty string is returned.
// There's no need for specific errors since the calling code will not care what
// exactly went wrong and just return an Unauthorized error/response.
func (r BearerTokenReader) Read(authHeader string) string {
	s := strings.Split(authHeader, " ")
	if s[0] != "Bearer" || s[1] == "" {
		return ""
	}
	return s[1]
}
