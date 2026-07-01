package main

import (
	"fmt"
	"net/http"
	"time"

	"movieapi.sanjayojha.dev/internal/data"
	"movieapi.sanjayojha.dev/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string       `json:"title"`
		Year    int          `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// validating data
	// Initialize a new Validator instance.
	v := validator.New()

	data.ValidateMovie(v, movie)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
		CreatedAt: time.Now(),
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	// if err != nil {
	// 	app.logger.Error(err.Error())
	// 	http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	// }
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
