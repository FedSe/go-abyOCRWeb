package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", app.homeGet)
	mux.HandleFunc("POST /", app.homePost)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.HandleFunc("GET /static/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	})

	return mux
}
