package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"main/modules"
)

func main() {
	godotenv.Load()

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

	r.Handle("/dashboard", modules.TokenCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(modules.Auth)
		dashboard := template.Must(template.ParseFiles("assets/dashboard.html"))
		dashboard.Execute(w, user)
	})))

	r.Handle("/", modules.TokenCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Context().Value("user")
		w.Write([]byte("Welcome to SmartWays"))
	})))

	r.HandleFunc("/maps", modules.GmapsProxyHandler)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), r); err != nil {
		panic(err)
	}
}
