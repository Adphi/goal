package goal

import (
	"fmt"
	"net/http"
	"reflect"
)

// QuerySupporter is the interface that return filtered results
// based on request parameters
type QuerySupporter interface {
	Query(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

func (g *Goal) queryHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if r, ok := resource.(QuerySupporter); ok {
			handler = r.Query
		} else if a, ok := g.resources[reflect.TypeOf(resource)]; ok && a.Query {
			handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
				return g.handleQuery(reflect.TypeOf(resource), r)
			}
		}

		renderJSON(rw, request, handler)
	}
}

// AddQueryResource allows model to support query based on request
// data, return filtered results back to client
func (g *Goal) AddQueryResource(resource interface{}, path string) {
	if a, ok := g.resources[reflect.TypeOf(resource)]; ok {
		a.Query = true
		g.resources[reflect.TypeOf(resource)] = a
	} else {
		g.resources[reflect.TypeOf(resource)] = ResourceACL{Query: true}
	}
	g.mux.Handle(path, g.queryHandler(resource))
}

// AddDefaultQueryPath allows model to support query based on request
// data, return filtered results back to client. The path is created
// base on struct name
func (g *Goal) AddDefaultQueryPath(resource interface{}) {
	queryPath := fmt.Sprintf("/query/%s/{query}", g.tableName(resource))
	g.AddQueryResource(resource, queryPath)
}
