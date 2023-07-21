package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
	

type MongoInstance struct {
	Client *mongo.Client
	Db	   *mongo.Database
} 

var mg MongoInstance

const dbName = "hrms"
const mongoURI = "mongodb://localhost:27017" + dbName

type Employee struct {
	ID		string
	Name	string
	Salary	float64
	Age		float64
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = MongoInstance{
		client,
		db,
	}
	return nil
}

func main() {
	app := fiber.New()

}