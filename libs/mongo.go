package libs

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"strconv"
)
import "goFrame/config"

type mongoSessionStoreStruct struct {
	_session *mgo.Session
}

var mongoSessionStore *mongoSessionStoreStruct

func init() {
	mongoSessionStore = new(mongoSessionStoreStruct)
	_ = GetMongoSession()
}

func newMongoSession() *mgo.Session {
	conf := config.GetConf()
	mongoUrl := fmt.Sprintf("mongodb://%s:%s", conf.Mongo.Host, strconv.Itoa(conf.Mongo.Port))
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		panic(fmt.Sprintf("mongo  连接失败 %s", err))
	}
	return session
}

// GetMongoSession ...
func GetMongoSession() *mgo.Session {
	if mongoSessionStore._session != nil {
		return mongoSessionStore._session
	} else {
		session := newMongoSession()
		mongoSessionStore._session = session
		return session
	}
}

// GetMongoDB ...
func GetMongoDB(name *string) *mgo.Database {
	session := GetMongoSession()
	DB := session.DB(*name)
	return DB
}

// GetMongoCol ...
func GetMongoCol(DBName *string, ColName *string) *mgo.Collection {
	DB := GetMongoDB(DBName)
	col := DB.C(*ColName)
	return col
}

// CloseMongoSession ...
func CloseMongoSession() {
	session := GetMongoSession()
	session.Close()
}
