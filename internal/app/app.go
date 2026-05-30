package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ArcaneCrowA/go-url-shortener/internal/handler"
	"github.com/ArcaneCrowA/go-url-shortener/internal/service"
	"github.com/ArcaneCrowA/go-url-shortener/internal/storage/mongodb"
	redisdb "github.com/ArcaneCrowA/go-url-shortener/internal/storage/redis"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Start() {
	var repo service.Repo

	switch os.Getenv("STORAGE_DRIVER") {
	case "mongodb":
		client, cleanup := connectMongoDB()
		defer cleanup()
		repo = mongodb.New(client)
	case "redis":
		rdb := connectRedis()
		defer rdb.Close()
		repo = redisdb.New(rdb)
	default:
		panic("not chosen")
	}

	srv := service.New(repo)
	hand := handler.New(srv)
	startServer(hand)
}

func connectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connection: %v", err)
	}

	return rdb
}

func connectMongoDB() (*mongo.Client, func()) {
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

	return client, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("mongo disconnection: %v", err)
		}
	}
}

func startServer(h *handler.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /{code}", h.Reroute)

	srv := http.Server{
		Handler: mux,
		Addr:    ":3000",
	}

	log.Fatal(srv.ListenAndServe())
}
