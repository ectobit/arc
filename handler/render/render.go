// Package render contans an abstraction over HTTP response rendering.
package render

import "net/http"

// Renderer defines abstraction methods of a HTTP server response renderer.
type Renderer interface {
	// Render renders HTTP response with body.
	Render(res http.ResponseWriter, statusCode int, body interface{})
	// Error renders HTTP response with error in body.
	Error(res http.ResponseWriter, statusCode int, message string)
}
