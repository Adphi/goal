package goal

import (
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

// Cacher defines a interface for fast key-value caching
type Cacher interface {
	Get(string, interface{}) error
	Set(string, interface{}) error
	Delete(string) error
	Exists(string) (bool, error)
	Close() error
}

func (g *Goal) registerCacher(cache Cacher) {
	logrus.Info("Registering cache")
	if g.cacher != nil {
		// Register Gorm callbacks
		if g.db != nil {
			logrus.Info("Registering DB cache callbacks")
			g.db.Callback().Create().After("gorm:after_create").Register("goal:cache_after_create", g.cache)
			g.db.Callback().Update().After("gorm:after_update").Register("goal:cache_after_update", g.cache)
			g.db.Callback().Query().After("gorm:after_query").Register("goal:cache_after_query", g.cache)
			g.db.Callback().Delete().Before("gorm:before_delete").Register("goal:uncache_after_delete", g.uncache)
		}
	}
}

func cacheKeyFromScope(scope *gorm.Scope) string {
	name := scope.TableName()
	id := scope.PrimaryKeyValue()
	key := defaultCacheKey(name, id)
	return key
}

// cacheKey defines by the struct or fallback
// to name:id format
func (g *Goal) cacheKey(resource interface{}) string {
	scope := g.db.NewScope(resource)
	return cacheKeyFromScope(scope)
}

// defaultCacheKey returns default format for redis key
func defaultCacheKey(name string, id interface{}) string {
	return fmt.Sprintf("%v:%v", name, id)
}

// uncache data from cache
func (g *Goal) uncache(scope *gorm.Scope) {
	logrus.Debug("Uncaching query")
	// Reload object before delete
	scope.DB().New().First(scope.Value)

	// Delete from cache
	key := cacheKeyFromScope(scope)
	g.cacher.Delete(key)
}

// cacher data to cache
func (g *Goal) cache(scope *gorm.Scope) {
	logrus.Debug("Caching query")
	key := cacheKeyFromScope(scope)
	g.cacher.Set(key, scope.Value)
}
