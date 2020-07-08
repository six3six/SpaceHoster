package login

import (
	"github.com/six3six/SpaceHoster/api/common"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type User struct {
	Login           Login
	EncodedPassword string
	Email           string
	Name            string
	Roles           []string
	Quota           common.Resource
	Database        *mongo.Database `bson:"-"`
}

type Login string

type Token struct {
	Token   string
	Login   Login
	LastUse time.Time
}
