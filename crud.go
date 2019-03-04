package goal

import (
	"fmt"
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
func (api *API) crudHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		switch request.Method {
		case http.MethodGet:
			if resource, ok := resource.(GetSupporter); ok {
				handler = resource.Get
				break
			}
			if a, ok := api.resources[reflect.TypeOf(resource)]; ok && a.Read {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return Read(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodPost:
			if resource, ok := resource.(PostSupporter); ok {
				handler = resource.Post
				break
			}
			if a, ok := api.resources[reflect.TypeOf(resource)]; ok && a.Create {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return Create(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodPut:
			if resource, ok := resource.(PutSupporter); ok {
				handler = resource.Put
				break
			}
			if a, ok := api.resources[reflect.TypeOf(resource)]; ok && a.Update {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return Update(reflect.TypeOf(resource), r)
				}
			}
		case http.MethodDelete:
			if resource, ok := resource.(DeleteSupporter); ok {
				handler = resource.Delete
				break
			}
			if a, ok := api.resources[reflect.TypeOf(resource)]; ok && a.Delete {
				handler = func(writer http.ResponseWriter, r *http.Request) (int, interface{}, error) {
					return Delete(reflect.TypeOf(resource), r)
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
func (api *API) AddCrudResource(resource interface{}, paths ...string) {
	for _, path := range paths {
		api.Mux().HandleFunc(path, api.crudHandler(resource))
	}
}

type Access struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
	Query  bool
}

func (api *API) AddResource(resource interface{}, access Access, paths ...string) {
	if api.resources == nil {
		api.resources = map[reflect.Type]Access{}
	}
	api.resources[reflect.TypeOf(resource)] = access
	for _, path := range paths {
		api.Mux().HandleFunc(path, api.crudHandler(resource))
	}
}

// AddDefaultCrudPaths adds default path for a resource.
// The default path is based on the struct name
func (api *API) AddDefaultCrudPaths(resource interface{}) {
	// Extract name of resource type
	name := TableName(resource)

	// Default path to interact with resource
	createPath := fmt.Sprintf("/%s", name)
	detailPath := fmt.Sprintf("/%s/{id:[a-zA-Z0-9]+}", name)

	api.AddCrudResource(resource, createPath, detailPath)
}
