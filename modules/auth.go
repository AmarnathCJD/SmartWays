package modules

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Auth struct {
	UserID      string `json:"userid"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Type        int    `json:"type"`
	AccessToken string `json:"accesstoken"`
}

const (
	ADMIN = iota + 1
	EMERGENCY
	USER
)

func convertType(t string) int {
	switch t {
	case "admin":
		return ADMIN
	case "emergency":
		return EMERGENCY
	case "user":
		return USER
	default:
		return USER
	}
}

func (a *Auth) Register() error {
	if a.IsUserExist() {
		return errors.New("User already exist")
	}

	a.UserID = genUserId()
	_, err := DB.Collection("users").InsertOne(context.Background(), a)
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) IsUserExist() bool {
	user, err := DB.Collection("users").Find(context.Background(), bson.M{"email": a.Email})
	if err != nil {
		return false
	}

	if user.Next(context.Background()) {
		return true
	}

	return false
}

func (a *Auth) Login() (Auth, error) {
	user := Auth{}
	err := DB.Collection("users").FindOne(context.Background(), bson.M{"email": a.Email, "password": a.Password}).Decode(&user)
	if err != nil {
		return user, errors.New("Invalid email or password")
	}

	return user, nil
}

func (a *Auth) GetUser() (Auth, error) {
	user := Auth{}
	err := DB.Collection("users").FindOne(context.Background(), bson.M{"email": a.Email}).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a *Auth) Update() error {
	_, err := DB.Collection("users").UpdateOne(context.Background(), bson.M{"email": a.Email}, bson.M{"$set": a})
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) Delete() error {
	_, err := DB.Collection("users").DeleteOne(context.Background(), bson.M{"email": a.Email})
	if err != nil {
		return err
	}

	return nil
}

func (a *Auth) GetUsers() ([]Auth, error) {
	users := []Auth{}
	cursor, err := DB.Collection("users").Find(context.Background(), bson.M{})
	if err != nil {
		return users, err
	}

	for cursor.Next(context.Background()) {
		user := Auth{}
		err := cursor.Decode(&user)
		if err != nil {
			return users, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (a *Auth) GetUserByID() (Auth, error) {
	user := Auth{}
	err := DB.Collection("users").FindOne(context.Background(), bson.M{"userid": a.UserID}).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a *Auth) GenUserToken() (string, error) {
	user, err := a.GetUser()
	if err != nil {
		return "", err
	}

	token, err := GenToken(user)
	DB.Collection("users").UpdateOne(context.Background(), bson.M{"email": a.Email}, bson.M{"$set": bson.M{"accesstoken": token}})

	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) VerifyToken(token string) (Auth, error) {
	user := Auth{}
	err := DB.Collection("users").FindOne(context.Background(), bson.M{"accesstoken": token}).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (a *Auth) ChangePassword(newPassword string) error {
	_, err := DB.Collection("users").UpdateOne(context.Background(), bson.M{"email": a.Email}, bson.M{"$set": bson.M{"password": newPassword}})
	if err != nil {
		return err
	}

	return nil
}

func GenToken(a Auth) (string, error) {
	currentTimeNano := time.Now().UnixNano()
	random := rand.Intn(1000)

	token := fmt.Sprintf("%s%d%d", a.UserID, currentTimeNano, random)
	return base64.StdEncoding.EncodeToString([]byte(token)), nil
}

func genUserId() string {
	length := 6
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	min := int64(math.Pow10(length - 1))
	max := int64(math.Pow10(length))

	id := random.Int63n(max-min) + min
	return fmt.Sprintf("SW%d", id)
}

type Alert struct {
	AlertID     string     `json:"alertid"`
	UserID      string     `json:"userid"`
	AlertType   int        `json:"alerttype"`
	Location    [2]float64 `json:"location"`
	Destination string     `json:"destination"`
	Severity    int        `json:"severity"`
	Clearance   bool       `json:"clearance"`
}

const (
	ACCIDENT = iota + 1
	FIRE
	POLICE
)

func convertAlertType(t string) int {
	switch t {
	case "accident":
		return ACCIDENT
	case "fire":
		return FIRE
	case "police":
		return POLICE
	default:
		return ACCIDENT
	}
}

func AddAlert(a Alert) error {
	_, err := DB.Collection("alerts").InsertOne(context.Background(), a)
	if err != nil {
		return err
	}

	return nil
}

func genAlertId() string {
	length := 6
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	min := int64(math.Pow10(length - 1))
	max := int64(math.Pow10(length))

	id := random.Int63n(max-min) + min
	return fmt.Sprintf("AL%d", id)
}

func GetAlerts() ([]Alert, error) {
	alerts := []Alert{}
	cursor, err := DB.Collection("alerts").Find(context.Background(), bson.M{})
	if err != nil {
		return alerts, err
	}

	for cursor.Next(context.Background()) {
		alert := Alert{}
		err := cursor.Decode(&alert)
		if err != nil {
			return alerts, err
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func RemoveAlert(alertID string) error {
	_, err := DB.Collection("alerts").DeleteOne(context.Background(), bson.M{"alertid": alertID})
	if err != nil {
		return err
	}

	return nil
}
