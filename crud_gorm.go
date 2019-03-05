// gorm_handlers provides basic methods to interact with
// database using GORM. https://github.com/jinzhu/gorm

package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
)

// read provides basic implementation to retrieve object
// based on request parameters
func (g *Goal) read(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter
	id, exists := vars["id"]
	if !exists {
		err := errors.New("id is required")
		return 400, nil, err
	}

	resource := newObjectWithType(rType)

	// Attempt to retrieve from redis first, if not exist, retrieve from
	// database and cacher it
	var err error
	if g.cacher != nil {
		name := g.tableName(resource)
		redisKey := defaultCacheKey(name, id)
		err = g.cacher.Get(redisKey, resource)
		if err == nil && resource != nil {
			// Check if resource is authorized
			err = g.CanPerform(resource, request, true)
			if err != nil {
				return 403, nil, err
			}

			return 200, resource, nil
		}
	}

	// Retrieve from database
	err = g.db.Where("id = ?", id).First(resource).Error
	if err != nil {
		return 500, nil, err
	}

	// Save to redis
	if g.cacher != nil {
		key := g.cacheKey(resource)
		g.cacher.Set(key, resource)
	}

	// Check if resource is authorized
	err = g.CanPerform(resource, request, true)
	if err != nil {
		return 403, nil, err
	}

	return 200, resource, nil
}

// create provides basic implementation to Create a record
// into the database
func (g *Goal) create(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	resource := newObjectWithType(rType)

	// Parse request body into resource
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(resource)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Save to database
	err = g.db.Create(resource).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, nil
}

// update provides basic implementation to update a record
// inside database
func (g *Goal) update(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id, exists := vars["id"]
	if !exists {
		err := errors.New("id is required")
		return 400, nil, err
	}

	resource := newObjectWithType(rType)

	// Parse request body into updatedObj
	updatedObj := newObjectWithType(rType)
	decoder := json.NewDecoder(request.Body)

	err := decoder.Decode(updatedObj)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Retrieve from database
	err = g.db.Where("id = ?", id).First(resource).Error
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Check permission
	err = g.CanPerform(resource, request, false)
	if err != nil {
		return 403, nil, err
	}

	// Check if this object support revision
	current, okCurrent := resource.(Revisioner)
	updated, okUpdated := updatedObj.(Revisioner)
	if okCurrent && okUpdated {
		if updated.CurrentRevision() == 0 {
			err = errors.New("revision is required")
			return 400, nil, err
		}

		if !CanMerge(current, updated) {
			err = errors.New("conflict")
			return 409, resource, err
		}

		updated.SetNextRevision()
	}

	// Save to database. Only update fields that is not blank or default values
	// http://jinzhu.me/gorm/curd.html#update
	err = g.db.Model(resource).Update(updatedObj).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, err
}

// delete provides basic implementation to delete a record inside
// a database
func (g *Goal) delete(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id, exists := vars["id"]
	if !exists {
		err := errors.New("id is required")
		return 400, nil, err
	}

	resource := newObjectWithType(rType)

	// Retrieve from database
	err := g.db.Where("id = ?", id).First(resource).Error
	if err != nil {
		return 500, nil, err
	}

	// Check permission
	err = g.CanPerform(resource, request, false)
	if err != nil {
		return 403, nil, err
	}

	// Delete record, if failed show 500 error code
	err = g.db.Delete(resource).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, nil, nil
}
