package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// Add the route for the /v1/movies endpoints.
	// Use the requireActivatedUser() middleware on /v1/movies** endpoints.
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser(app.deleteMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser(app.listMovieHandler))

	// Add the route for the /v1/users endpoints.
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// genrate token
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Wrap the router with the panic recovery middleware.
	// Use the authenticate() middleware on all requests.
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
