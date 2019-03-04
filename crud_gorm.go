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
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql" // Driver for mysql
	_ "github.com/lib/pq"              // Driver for postgres
	_ "github.com/mattn/go-sqlite3"    // Driver for sqlite
)

// Global variable to interact with database
var db *gorm.DB

// InitGormDb initializes global variable db
func InitGormDb(newDb *gorm.DB) {
	db = newDb
}

// DB returns global variable db
func DB() *gorm.DB {
	return db
}

// Read provides basic implementation to retrieve object
// based on request parameters
func Read(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

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
	// database and cache it
	var err error
	if SharedCache != nil {
		name := TableName(resource)
		redisKey := DefaultCacheKey(name, id)
		err = SharedCache.Get(redisKey, resource)
		if err == nil && resource != nil {
			// Check if resource is authorized
			err = CanPerform(resource, request, true)
			if err != nil {
				return 403, nil, err
			}

			return 200, resource, nil
		}
	}

	// Retrieve from database
	err = db.Where("id = ?", id).First(resource).Error
	if err != nil {
		return 500, nil, err
	}

	// Save to redis
	if SharedCache != nil {
		key := CacheKey(resource)
		SharedCache.Set(key, resource)
	}

	// Check if resource is authorized
	err = CanPerform(resource, request, true)
	if err != nil {
		return 403, nil, err
	}

	return 200, resource, nil
}

// Create provides basic implementation to Create a record
// into the database
func Create(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	resource := newObjectWithType(rType)

	// Parse request body into resource
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(resource)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Save to database
	err = db.Create(resource).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, nil
}

// Update provides basic implementation to update a record
// inside database
func Update(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

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
	err = db.First(resource, id).Error
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Check permission
	err = CanPerform(resource, request, false)
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
	err = db.Model(resource).Update(updatedObj).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, err
}

// Delete provides basic implementation to delete a record inside
// a database
func Delete(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

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
	err := db.First(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	// Check permission
	err = CanPerform(resource, request, false)
	if err != nil {
		return 403, nil, err
	}

	// Delete record, if failed show 500 error code
	err = db.Delete(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, nil, nil
}
