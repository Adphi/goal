package goal

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

// To shorten the code, define a type
type simpleResponse func(http.ResponseWriter, *http.Request) (int, interface{}, error)

// tableName returns table name for the resource
func (g *Goal) tableName(resource interface{}) string {
	// Extract name of resource type
	name := g.db.NewScope(resource).TableName()
	return name
}

// Error message should be a json object, with error message
// and any optional data
func getErrorString(data interface{}, err error) string {
	errMap := map[string]interface{}{
		"message": err.Error(),
	}

	if data != nil {
		errMap["data"] = data
	}

	errByte, marshalErr := json.Marshal(errMap)
	if marshalErr != nil {
		errMap["data"] = marshalErr.Error()
		errByte, _ = json.Marshal(errMap)
	}

	return string(errByte)
}

// Write response back to client
func renderJSON(rw http.ResponseWriter, request *http.Request, handler simpleResponse) {
	if handler == nil {
		http.Error(rw, http.ErrNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	code, data, err := handler(rw, request)

	if err != nil {
		http.Error(rw, getErrorString(data, err), code)
		return
	}

	var content []byte
	content, err = json.Marshal(data)
	if err != nil {
		http.Error(rw, getErrorString(nil, err), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	rw.Write(content)
}

// RegisterModel initializes default routes for a model
func (g *Goal) RegisterModel(resource interface{}, access ResourceACL) {
	logrus.Infof("Registering model : %s", g.tableName(resource))
	g.db.AutoMigrate(resource)
	if g.resources == nil {
		g.resources = map[reflect.Type]ResourceACL{}
	}
	g.resources[reflect.TypeOf(resource)] = access
	g.AddDefaultCrudPaths(resource)
	g.AddDefaultQueryPath(resource)
}

// dynamicSlice creates a slice with element with resource type
// Copied from http://stackoverflow.com/a/25386460/622510 (Thanks @nemo)
func dynamicSlice(resource interface{}) interface{} {
	rType := reflect.TypeOf(resource)

	// Create a slice to begin with
	slice := reflect.MakeSlice(reflect.SliceOf(rType), 0, 0)

	// Create a pointer to a slice value and set it to the slice
	x := reflect.New(slice.Type())
	x.Elem().Set(slice)

	return x.Interface()
}

func newObjectWithType(t reflect.Type) interface{} {
	if reflect.TypeOf(t).Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface()
}
