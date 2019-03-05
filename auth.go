package goal

import "net/http"

// Registerer register a new user to system
type Registerer interface {
	Register(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// Loginer authenticates user into the system
type Loginer interface {
	Login(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// Logouter clear sessions and log user out
type Logouter interface {
	Logout(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

func (g *Goal) registerHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Registerer); ok {
			handler = resource.Register
		}

		renderJSON(rw, request, handler)
	}
}

func (g *Goal) loginHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Loginer); ok {
			handler = resource.Login
		}

		renderJSON(rw, request, handler)
	}
}

func (g *Goal) logoutHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Logouter); ok {
			handler = resource.Logout
		}

		renderJSON(rw, request, handler)
	}
}

// AddRegisterPath let user to register into a system
func (g *Goal) AddRegisterPath(resource interface{}, path string) {
	g.mux.Handle(path, g.registerHandler(resource))
}

// AddLoginPath let user login to system
func (g *Goal) AddLoginPath(resource interface{}, path string) {
	g.mux.Handle(path, g.loginHandler(resource))
}

// AddLogoutPath let user logout from the system
func (g *Goal) AddLogoutPath(resource interface{}, path string) {
	g.mux.Handle(path, g.logoutHandler(resource))
}

// AddDefaultAuthPaths route request to the model which implement
// authentications
func (g *Goal) AddDefaultAuthPaths(resource interface{}) {
	g.mux.Handle("/auth/register", g.registerHandler(resource))
	g.mux.Handle("/auth/login", g.loginHandler(resource))
	g.mux.Handle("/auth/logout", g.logoutHandler(resource))
}
