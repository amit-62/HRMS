package main

import (
	"context"
	"fmt"
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
const mongoURI = "mongodb://localhost:27017/" + dbName

// type Employee struct {
// 	ID		string		`json:"id,omitempty" bson:"_id,omitempty"`
// 	Name	string		`json:"name"`
// 	Salary	float64		`json:"salary"`
// 	Age		float64		`json:"age"`
// }

type Employee struct {
    ID       string 			`json:"_id,omitempty" bson:"_id,omitempty"`
    Name     string             `json:"name"`
    Salary   int                `json:"salary"`
    Age      int                `json:"age"`
}

func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Println("connected")
	db := client.Database(dbName)

	mg = MongoInstance{
		Client: client,
		Db: db,
	}
	return nil
}

func main() {
	if err := Connect(); err != nil{
		log.Fatal(err)
	}
	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var employees []Employee = make([]Employee, 0)
		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(employees);
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		employee := new(Employee)
		fmt.Println("Request Body:", string(c.Body()))
		if err := c.BodyParser(&employee); err != nil {
			fmt.Println("Error parsing request body:", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse request body: " + err.Error(),
			})
		}
		employee.ID = ""

		insertionResult, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployee := &Employee{}
		createdRecord.Decode(createdEmployee)

		return c.Status(201).JSON(createdEmployee)

	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			return c.SendStatus(400)
		}

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{{Key: "_id", Value: employeeID}}
		update := bson.D{
			{Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "age", Value: employee.Age},
					{Key: "salary", Value: employee.Salary},
				},
			},
		}

		err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}

		employee.ID = idParam

		return c.Status(200).JSON(employee)

	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeId, err := primitive.ObjectIDFromHex(idParam)

		if err!= nil{
			return c.SendStatus(404)
		}

		query := bson.D{{Key:"_id", Value:employeeId}}
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), query)

		if err!= nil{
			return c.SendStatus(500)
		}

		if result.DeletedCount <1 {
			return c.SendStatus(404)
		}

		return c.Status(200).JSON("record deleted")
	})

	log.Fatal(app.Listen(":3001"))

}