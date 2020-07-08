package login

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"log"
	"time"
)

type Service struct {
	Database *mongo.Database
}

func (service *Service) CleanTokens() {
	tokens := service.Database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	res, _ := tokens.DeleteMany(ctx, bson.M{"lastuse": bson.M{"$lt": time.Unix(time.Now().Unix()-int64(TokenDuration.Seconds()), 0)}})
	if res != nil {
		log.Printf("Clear %d token(s)", res.DeletedCount)
	}
}

func (service *Service) CleanToken(token string) error {
	tokens := service.Database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	_, err := tokens.DeleteOne(ctx, bson.M{"token": token})
	return err
}

func (service *Service) CheckToken(token string) (*User, error) {
	tokens := service.Database.Collection("tokens")
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)

	var tokenInfos Token
	err := tokens.FindOne(ctx, bson.M{"token": token}).Decode(&tokenInfos)
	if err != nil {
		return nil, err
	}

	logins := service.Database.Collection("logins")
	var user User
	err = logins.FindOne(ctx, bson.M{"login": tokenInfos.Login}).Decode(&user)
	if err != nil {
		return nil, err
	}

	_, err = logins.UpdateOne(ctx, bson.M{"login": tokenInfos.Login}, bson.M{"$set": bson.M{"lastuse": time.Now()}})
	if err != nil {
		return nil, err
	}
	user.Database = service.Database
	return &user, nil
}
