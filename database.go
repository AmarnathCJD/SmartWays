package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB   = initDB()
	opts = options.Update().SetUpsert(true)
	LOG  = log.New(os.Stdout, "SmartWays: ", log.LstdFlags|log.Lshortfile)
)

func initDB() *mongo.Database {
	godotenv.Load()

	clientOptions := options.Client().
		ApplyURI(os.Getenv("MONGO_URI"))
	db, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	DB := db.Database("smartways")

	LOG.Println("Connected to MongoDB!!")
	return DB
}
