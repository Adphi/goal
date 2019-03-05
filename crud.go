package goal

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

// GetSupporter is the interface that provides the Get
// method a resource must support to receive HTTP GETs.
type GetSupporter interface {
	Get(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// PostSupporter is the interface that provides the Post
// method a resource must support to receive HTTP POSTs.
type PostSupporter interface {
	Post(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// PutSupporter is the interface that provides the Put
// method a resource must support to receive HTTP PUTs.
type PutSupporter interface {
	Put(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// DeleteSupporter is the interface that provides the Delete
// method a resource must support to receive HTTP DELETEs.
type DeleteSupporter interface {
	Delete(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// HeadSupporter is the interface that provides the Head
// method a resource must support to receive HTTP HEADs.
type HeadSupporter interface {
	Head(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// PatchSupporter is the interface that provides the Patch
// method a resource must support to receive HTTP PATCHs.
type PatchSupporter interface {
	Patch(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// Route request to correct handler and write result back to client
func (g *Goal) crudHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		switch request.Method {
		case http.MethodGet:
			if resource, ok := resource.(GetSupporter); ok {
				handler = resource.Get
				break
			}
			if a, ok := g.resources[reflect.TypeOf(resource)]; ok && a.Read {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return g.read(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodPost:
			if resource, ok := resource.(PostSupporter); ok {
				handler = resource.Post
				break
			}
			if a, ok := g.resources[reflect.TypeOf(resource)]; ok && a.Create {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return g.create(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodPut:
			if resource, ok := resource.(PutSupporter); ok {
				handler = resource.Put
				break
			}
			if a, ok := g.resources[reflect.TypeOf(resource)]; ok && a.Update {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return g.update(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodDelete:
			if resource, ok := resource.(DeleteSupporter); ok {
				handler = resource.Delete
				break
			}
			if a, ok := g.resources[reflect.TypeOf(resource)]; ok && a.Delete {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return g.delete(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodHead:
			if resource, ok := resource.(HeadSupporter); ok {
				handler = resource.Head
			}
		case http.MethodPatch:
			if resource, ok := resource.(PatchSupporter); ok {
				handler = resource.Patch
			}
		}

		renderJSON(rw, request, handler)
	}
}

// AddCrudResource adds a new resource to an API. The API will route
// requests that match one of the given paths to the matching HTTP
// method on the resource.
func (g *Goal) AddCrudResource(resource interface{}, paths ...string) {
	for _, path := range paths {
		g.mux.HandleFunc(path, g.crudHandler(resource))
	}
}

type ResourceACL struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
	Query  bool
}

func AllACL() ResourceACL {
	return ResourceACL{
		Read:   true,
		Create: true,
		Update: true,
		Delete: true,
		Query:  true,
	}
}

func (g *Goal) AddResource(resource interface{}, access ResourceACL, paths ...string) {
	logrus.Infof("Adding resource : %s", g.tableName(resource))
	if g.resources == nil {
		g.resources = map[reflect.Type]ResourceACL{}
	}
	g.resources[reflect.TypeOf(resource)] = access
	for _, path := range paths {
		g.mux.HandleFunc(path, g.crudHandler(resource))
	}
}

// AddDefaultCrudPaths adds default path for a resource.
// The default path is based on the struct name
func (g *Goal) AddDefaultCrudPaths(resource interface{}) {
	logrus.Debugf("Adding default paths for model : %s", g.tableName(resource))
	// Extract name of resource type
	name := g.tableName(resource)

	// Default path to interact with resource
	createPath := fmt.Sprintf("/%s", name)
	detailPath := fmt.Sprintf("/%s/{id:[a-zA-Z0-9]+}", name)

	g.AddCrudResource(resource, createPath, detailPath)
}
