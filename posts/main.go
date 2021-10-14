package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type UserPosted struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Caption  string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL string             `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	Posted   time.Time          `json:"posted,omitempty" bson:"posted,omitempty"`
}

func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var posts UserPosted
	json.NewDecoder(request.Body).Decode(&posts)
	posts.Posted = time.Now()
	collection := client.Database("instagram").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, posts)
	json.NewEncoder(response).Encode(result)

}

func GetUserpostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var posts UserPosted
	collection := client.Database("instagram").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, UserPosted{ID: id}).Decode(&posts)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(posts)
}

func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var userspost []UserPosted
	collection := client.Database("instagram").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var posts UserPosted
		cursor.Decode(&posts)
		userspost = append(userspost, posts)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(userspost)
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017") //This will connect to mongoDB
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/post", CreatePostEndpoint).Methods("POST")  //This will create user
	router.HandleFunc("/post", GetUserpostEndpoint).Methods("GET")  //This will deploy all created users
	router.HandleFunc("/post/{id}", GetPostEndpoint).Methods("GET") //This will deploy the specific ID
	http.ListenAndServe(":12345", router)                           //URL port can be anything
}
