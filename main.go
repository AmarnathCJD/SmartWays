package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"main/modules"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to SmartWays"))
	})

	// API routes
	r.Post("/api/login", modules.LoginHandler)
	r.Post("/api/register", modules.RegisterHandler)

	// HTML routes
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/login.html")
	})
	r.Get("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/register.html")
	})

	r.Handle("/", modules.TokenCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Context().Value("user")
		w.Write([]byte("Welcome to SmartWays"))
	})))

	http.ListenAndServe(":3000", r)
}
