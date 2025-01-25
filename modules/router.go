package modules

import (
	"context"
	"fmt"
	"net/http"
)

func TokenCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		token := ""
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				token = cookie.Value
			}
		}
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Set-Cookie", fmt.Sprintf("token=%s", userToken))

	w.Write([]byte(`{"message": "Login successful", "token": "` + userToken + `", "user_id": "` + user.UserID + `"}`))
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

}
