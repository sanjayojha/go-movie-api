package data

import "time"

type Movie struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Genres    []string  `json:"genres,omitempty"` //omtiempty will remove empty slices or map
	Year      int       `json:"year,omitzero"`
	Runtime   int       `json:"runtime,omitzero"` // movie runtime in minutes
	Version   int       `json:"version"`          // start with 1, updated when move info updated
	CreatedAt time.Time `json:"-"`
}
