package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func TokenCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		a := Auth{}

		user, err := a.VerifyToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.FormValue("email")
	password := r.FormValue("password")

	auth := Auth{
		Email:    email,
		Password: password,
	}

	user, err := auth.Login()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userToken, _ := user.GenUserToken()

	writeJSON(w, map[string]interface{}{
		"message": "Login successful",
		"token":   userToken,
		"user_id": user.UserID,
	})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")
	userType := r.FormValue("userType")

	fmt.Println(email, password, name, userType)

	auth := Auth{
		Email:    email,
		Password: password,
		Name:     name,
		Type:     convertType(userType),
	}

	err := auth.Register()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"message": "User created successfully",
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
