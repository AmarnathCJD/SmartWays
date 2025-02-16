package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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

func GetJunctionHandler(w http.ResponseWriter, r *http.Request) {
	jid := r.URL.Query().Get("jid")
	jidInt, err := strconv.Atoi(jid)
	if err != nil {
		http.Error(w, `{"message": "Invalid junction ID"}`, http.StatusBadRequest)
		return
	}
	for _, j := range Junctions {
		if j.JnID == jidInt {
			w.Header().Set("Content-Type", "application/json")
			jn, _ := json.Marshal(j)
			w.Write(jn)
			return
		}
	}
	http.Error(w, `{"message": "Junction not found"}`, http.StatusNotFound)
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

func AddAlertHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var resp struct {
		UserId    string `json:"user_id"`
		AlertType string `json:"type"`
		Location  struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
		Severity  string `json:"severity"`
		Dest      string `json:"destination"`
		Clearance bool   `json:"requiresClearance"`
	}

	err := json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alert := Alert{
		AlertID:     genAlertId(),
		UserID:      resp.UserId,
		AlertType:   convertAlertType(resp.AlertType),
		Location:    [2]float64{resp.Location.Latitude, resp.Location.Longitude},
		Destination: resp.Dest,
		Severity:    convertInt(resp.Severity),
		Clearance:   resp.Clearance,
	}

	broadcast <- TrafficEvent{Type: "alert", Direction: alert.Destination, Color: resp.AlertType, CountAllowed: alert.Severity}

	err = AddAlert(alert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"message": "Alert added successfully", "alert_id": "` + alert.AlertID + `"}`))
}

func RemoveAlertHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	alertId := r.FormValue("alert_id")
	err := RemoveAlert(alertId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"message": "Alert removed successfully"}`))
}

func GetAlertsHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := GetAlerts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	alertsJson, _ := json.Marshal(alerts)
	w.Header().Set("Content-Type", "application/json")
	w.Write(alertsJson)
}
