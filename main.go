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
	r.Post("/api/spawn", modules.SpawnRequestHandler)
	r.Post("/api/toggle-spawn", modules.HandleRandomToggle)
	r.Post("/api/toggle-auto", modules.HandleAutoToggle)
	r.Post("/api/clear-vehicles", modules.ClearVehiclesHandler)

	r.Get("/api/jn", modules.GetJunctionHandler)

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

	r.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/index.html")
	})
	r.HandleFunc("/spawn", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/spawn.html")
	})
	r.HandleFunc("/emergency", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/emergency.html")
	})
	r.HandleFunc("/ws", modules.HandleWSConnections)

	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	r.Handle("/", modules.TokenCheck(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Context().Value("user")
		w.Write([]byte("Welcome to SmartWays"))
	})))

	r.HandleFunc("/maps", modules.GmapsProxyHandler)

	// ALERTS
	r.Get("/api/alerts", modules.GetAlertsHandler)
	r.Post("/api/alerts", modules.AddAlertHandler)
	r.Post("/api/delalert", modules.RemoveAlertHandler)

	// sse
	go modules.StartTrafficSimulation()
	go modules.BroadcastTrafficUpdates()

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), r); err != nil {
		panic(err)
	}
}
