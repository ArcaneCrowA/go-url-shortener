package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoData struct {
	Code string `bson:"code"`
	Site string `bson:"site"`
}

type Repo struct {
	client *mongo.Client
}

func New(client *mongo.Client) *Repo {
	return &Repo{
		client: client,
	}
}

func (r *Repo) InsertRecord(code, site string) error {
	coll := r.client.Database("db").Collection("websites")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	website := MongoData{
		Code: code,
		Site: site,
	}
	_, err := coll.InsertOne(ctx, website)
	if err != nil {
		return errors.New("can't insert to collection")
	}
	return nil
}

func (r *Repo) GetRecord(code string) (string, error) {
	coll := r.client.Database("db").Collection("websites")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var data MongoData
	err := coll.FindOne(ctx, bson.M{"code": code}).Decode(&data)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("record not found")
		} else {
			return "", errors.New("internal error")
		}
	}

	return data.Site, nil
}
