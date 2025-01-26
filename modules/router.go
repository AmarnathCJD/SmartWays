package modules

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

func GmapsProxyHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "https://maps.googleapis.com/maps/api/js?key="+os.Getenv("GMAPS_API_KEY"), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
