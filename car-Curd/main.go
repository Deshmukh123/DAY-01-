package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Car struct {
	Name  string `json:"name" bson:"name"`
	Model string `json:"model" bson:"model"`
}

var client *mongo.Client
var collection *mongo.Collection

func main() {
	// Set up  MongoDB client
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	// Initialized the collection
	collection = client.Database("car_db").Collection("cars")

	r := gin.Default()

	r.POST("/cars", createCar)
	r.GET("/cars", getCars)
	r.GET("/cars/:name", getCarByName)
	r.PUT("/cars/:name", updateCar)
	r.DELETE("/cars/:name", deleteCar)

	r.Run() //port 8080
}

func createCar(c *gin.Context) {
	var newCar Car
	if err := c.ShouldBindJSON(&newCar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := collection.InsertOne(context.TODO(), newCar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert car"})
		return
	}
	c.JSON(http.StatusCreated, newCar)
}

func getCars(c *gin.Context) {
	var cars []Car
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cars"})
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var car Car
		if err := cursor.Decode(&car); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode car"})
			return
		}
		cars = append(cars, car)
	}
	c.JSON(http.StatusOK, cars)
}

func getCarByName(c *gin.Context) {
	name := c.Param("name")
	var car Car
	err := collection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&car)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "car not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve car"})
		}
		return
	}
	c.JSON(http.StatusOK, car)
}

func updateCar(c *gin.Context) {
	name := c.Param("name")
	var updatedCar Car
	if err := c.ShouldBindJSON(&updatedCar); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := collection.UpdateOne(context.TODO(), bson.M{"name": name}, bson.M{"$set": updatedCar})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update car"})
		return
	}
	c.JSON(http.StatusOK, updatedCar)
}

func deleteCar(c *gin.Context) {
	name := c.Param("name")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"name": name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete car"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "car deleted"})
}
