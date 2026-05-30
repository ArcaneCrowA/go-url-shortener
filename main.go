package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type App struct {
	client *mongo.Client
}

func newApp() *App {
	return &App{}
}

func main() {
	app := newApp()
	cleanup := app.connectDB()
	defer cleanup()

	app.startServer()
}

func (app *App) connectDB() func() {
	uri := os.Getenv("MONGODB_URL")
	if uri == "" {
		log.Fatal("set your MONGODB_URL variable")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connection: %v", err)
	}

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	coll := client.Database("db").Collection("websites")
	coll.Indexes().CreateOne(context.Background(), indexModel)

	app.client = client

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("mongo disconnection: %v", err)
		}
	}
}

func (app *App) startServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", app.shorten)
	mux.HandleFunc("GET /{code}", app.reroute)

	srv := http.Server{
		Handler: mux,
		Addr:    ":3000",
	}

	log.Fatal(srv.ListenAndServe())
}

type Data struct {
	Website string `json:"website"`
}

type MongoData struct {
	Code string `bson:"code"`
	Site string `bson:"site"`
}

func (app *App) shorten(w http.ResponseWriter, r *http.Request) {
	var data Data

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("can't read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Printf("can't read body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code, err := uuid.NewUUID()
	if err != nil {
		log.Printf("can't create uuid: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	coll := app.client.Database("db").Collection("websites")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortCode := code.String()[:6]

	website := MongoData{
		Code: shortCode,
		Site: data.Website,
	}
	_, err = coll.InsertOne(ctx, website)
	if err != nil {
		log.Printf("insert error : %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("stored: code=%q site=%q", shortCode, data.Website)
	w.Write([]byte(shortCode))
}

func (app *App) reroute(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	log.Printf("reroute called with code=%q", code)

	coll := app.client.Database("db").Collection("websites")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var data MongoData
	err := coll.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	if err != nil {
		log.Printf("FindOne error: %v", err)
		if err == mongo.ErrNoDocuments {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, data.Site, http.StatusPermanentRedirect)
}
