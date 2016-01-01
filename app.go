package main

import (
	"fmt"
	"github.com/aacanakin/hn"
	"github.com/aacanakin/hnr/controllers"
	"github.com/aacanakin/hnr/resources"
	"github.com/aacanakin/hnr/util"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/brandfolder/gin-gorelic"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"net/http"
	"time"
)

const (
	version = "0.0.0"
)

func main() {

	// initialize viper
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./.")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("%s\n", err))
	}

	// set mode
	if viper.GetString("app.env") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// initialize hn
	hn := hn.NewClient(&http.Client{
		Timeout: time.Duration(5 * time.Second),
	})

	// initialize memcache
	mc := memcache.New(fmt.Sprintf("%s:%d",
		viper.GetString("memcache.host"),
		viper.GetInt("memcache.port")),
	)

	mongoUrl := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		viper.GetString("mongo.user"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.host"),
		viper.GetInt("mongo.port"),
		viper.GetString("mongo.db"),
	)

	// initialize mgo
	mgo, err := mgo.Dial(mongoUrl)
	defer mgo.Close()

	if err != nil {
		panic(err)
	}

	util.CleanLocks()

	app := gin.Default()

	StoryResource := &resources.StoryResource{mc, mgo}

	Controller := controllers.Controller{App: app}
	StoryController := &controllers.StoryController{Controller, StoryResource, hn}
	StatusController := &controllers.StatusController{}

	if viper.GetString("app.env") == "prod" {
		gorelic.InitNewrelicAgent(viper.GetString("newrelic.license_key"), viper.GetString("app.name"), true)
		app.Use(gorelic.Handler)
	}

	// define routes
	app.GET("/status", StatusController.Ping)
	app.PUT("/stories/:type", StoryController.Cache)
	app.GET("/stories/:type", StoryController.Get)
	app.GET("/comments/:storyId", StoryController.GetComments)

	host := viper.GetString("http.host")
	port := viper.GetInt("http.port")
	env := viper.GetString("app.env")

	fmt.Printf("App started on env=%s host=%s port=%d version=%s\n",
		env, host, port, version)

	fmt.Println(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), app))
}
