package goal

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"net/http/httptest"
)

type testuser struct {
	ID       uint `gorm:"primary_key"`
	Username string
	Password string
	Name     string
	Age      int
	Rev      int64
}

type article struct {
	ID     uint `gorm:"primary_key"`
	Author *testuser
	Title  string
	Permission
}

var (
	g              *Goal
	testServer     *httptest.Server
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func setup() {
	// We do not setup db in order to sqlite3 in memory (default)
	var options []Option
	// Setup redis
	redisCache, err := NewRedisCache(*redisAddress, *maxConnections)
	if err == nil {
		options = append(options, WithCache(redisCache))
	}
	options = append(options,
		//Initialize db
		WithDBAddress("sqlite3", ":memory:"),
		// Set Singular tables
		WithDBOptions(func(db *gorm.DB) error {
			db.SingularTable(true)
			return nil
		}),
		// Initialize session
		WithSessionStore([]byte("something-very-secret")),
	)

	g, err = NewGoal(options...)
	if err != nil {
		logrus.Fatal(err)
	}

	// Initialize resource
	models := []interface{}{&testuser{}, &article{}}

	// Add default path
	for _, model := range models {
		g.RegisterModel(model, ResourceACL{
			Create: true,
			Read:   true,
			Update: true,
			Delete: true,
			Query:  true,
		})
	}

	user := &testuser{}
	g.SetUserModel(user)
	g.AddDefaultAuthPaths(user)

	testServer = httptest.NewServer(g.mux)
}

func tearDown() {
	if g != nil {
		if err := g.Close(); err != nil {
			logrus.Error(err)
		}
	}
	if testServer != nil {
		testServer.Close()
	}
}

func userURL() string {
	return fmt.Sprint(testServer.URL, "/testuser")
}

func idURL(id interface{}) string {
	return fmt.Sprint(testServer.URL, "/testuser/", id)
}
