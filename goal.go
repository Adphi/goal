package goal

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"gitlab.bertha.cloud/partitio/pqstream"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
)

type Goal struct {
	ctx context.Context
	c   *conf

	db       *gorm.DB
	cacher   Cacher
	mux      *mux.Router
	session  sessions.Store
	streamer *pqstream.Streamer

	resources map[reflect.Type]ResourceACL
	userType  reflect.Type
}

type conf struct {
	address     string
	dbDriver    string
	dbAddress   string
	sessionName string
	sessionKey  string
}

type Option func(*Goal) error

func NewGoal(options ...Option) (*Goal, error) {
	g := &Goal{resources: map[reflect.Type]ResourceACL{}, c: &conf{
		// sessionName is default name for user session
		sessionName: "goal.UserSessionName",
		// sessionKey is default key for user object
		sessionKey: "goal.UserSessionKey",
	}}

	// Create router
	g.mux = mux.NewRouter()

	// Set options
	for _, o := range options {
		if err := o(g); err != nil {
			logrus.Error(err)
			return nil, err
		}
	}
	// Init context
	g.ctx = BackgroundWithSignals(g.ctx)

	// Check db
	if g.db == nil {
		logrus.Warn("Goal need a database in order to work properly. Using sqlite in memory")
		var err error
		g.db, err = gorm.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, err
		}
	}

	// Create session if not set
	if g.session == nil {
		g.session = sessions.NewCookieStore([]byte("you-should-set-the-key-yourself"))
	}

	return g, nil
}

func (g *Goal) Mux() *mux.Router {
	return g.mux
}

func (g *Goal) Close() error {
	var errs []string
	if g.streamer != nil {
		if err := g.streamer.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if g.db != nil {
		if err := g.db.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if g.cacher != nil {
		if err := g.cacher.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ". "))
	}
	return nil
}

func (g *Goal) Run() error {
	return http.ListenAndServe(g.c.address, g.mux)
}

func WithContext(ctx context.Context) Option {
	return func(goal *Goal) error {
		if ctx == nil {
			return ErrNilContext
		}
		goal.ctx = BackgroundWithSignals(ctx)
		go func() {
			for {
				select {
				case <-ctx.Done():
					logrus.Info("Context cancelled")
					logrus.Info("Shutting down")
				}
			}
		}()
		return nil
	}
}

func WithAddress(address string) Option {
	return func(goal *Goal) error {
		goal.c.address = address
		return nil
	}
}

func WithCache(cache Cacher) Option {
	return func(goal *Goal) error {
		if cache == nil {
			return ErrNilCache
		}
		goal.registerCacher(cache)
		return nil
	}
}

func WithDBAddress(driver, address string) Option {
	return func(goal *Goal) error {
		if driver == "" {
			return ErrEmptyDBDriver
		}
		if address == "" {
			return ErrEmptyDBAddress
		}
		goal.c.dbDriver = driver
		goal.c.dbAddress = address

		db, err := gorm.Open(driver, address)
		if err != nil {
			return err
		}
		goal.db = db
		return nil
	}
}

func WithDBOptions(callback func(db *gorm.DB) error) Option {
	return func(goal *Goal) error {
		return callback(goal.db)
	}
}

func WithLiveQueries() Option {
	return func(goal *Goal) error {
		return nil
	}
}

func WithSessionName(name string) Option {
	return func(goal *Goal) error {
		if name != "" {
			goal.c.sessionName = name
		}
		return nil
	}
}

func WithSessionKey(key string) Option {
	return func(goal *Goal) error {
		if key != "" {
			goal.c.sessionKey = key
		}
		return nil
	}
}

func WithSessionStore(keyPairs ...[]byte) Option {
	return func(goal *Goal) error {
		goal.session = sessions.NewCookieStore(keyPairs...)
		return nil
	}
}

// BackgroundWithSignals returns a Context that will be
// canceled with the process receives a SIGINT signal.
// This function starts a goroutine and listens for signals.
func BackgroundWithSignals(p context.Context) context.Context {
	if p == nil {
		p = context.Background()
	}
	ctx, cancel := context.WithCancel(p)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Reset(os.Interrupt)
		cancel()
	}()
	return ctx
}
