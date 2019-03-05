package goal

import (
	"errors"
	"net/http"
	"reflect"
)

// SetUserModel lets goal which model act as user
func (g *Goal) SetUserModel(user interface{}) {
	g.userType = reflect.TypeOf(user).Elem()
}

// getUserResource returns a new variable based on reflection
// e.g user := &User{}
func (g *Goal) getUserResource() (interface{}, error) {
	if g.userType == nil {
		return nil, errors.New("User model was not registered")
	}

	return reflect.New(g.userType).Interface(), nil
}

// SetUserSession sets current user to session
func (g *Goal) setUserSession(w http.ResponseWriter, req *http.Request, user interface{}) error {
	session, err := g.session.Get(req, g.c.sessionName)
	if err != nil {
		return err
	}

	scope := g.db.NewScope(user)

	// Set some session values.
	session.Values[g.c.sessionKey] = scope.PrimaryKeyValue()

	// Save it before we write to the response/return from the handler.
	err = session.Save(req, w)
	return err
}

// GetCurrentUser returns current user based on the request header
func (g *Goal) getCurrentUser(req *http.Request) (interface{}, error) {
	session, err := g.session.Get(req, g.c.sessionName)
	if err != nil {
		return nil, err
	}

	userID, ok := session.Values[g.c.sessionKey]
	if !ok {
		return nil, errors.New("empty session")
	}

	var user interface{}
	user, err = g.getUserResource()
	if err != nil {
		return nil, err
	}

	// Load user from cacher or from database
	exists := false
	if g.cacher != nil {
		cacheKey := defaultCacheKey(g.tableName(user), userID)
		exists, err = g.cacher.Exists(cacheKey)
		if err == nil && exists {
			err = g.cacher.Get(cacheKey, user)

			if err == nil {
				return user, nil
			}
		}
	}

	// If data not exists in Redis, load from database
	if !exists {
		err = g.db.First(user, userID).Error
		return user, err
	}

	return nil, errors.New("invalid session data")
}

// clearUserSession removes the current user from session
func clearUserSession(w http.ResponseWriter, req *http.Request) error {
	http.SetCookie(w, nil)
	return nil
}
