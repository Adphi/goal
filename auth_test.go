package goal

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Setup methods to conform to auth interfaces
func (user *testuser) Register(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := g.RegisterWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Login(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := g.LoginWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Logout(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	g.HandleLogout(w, req)
	return 200, nil, nil
}

func TestAuth(t *testing.T) {
	setup()
	defer tearDown()

	recorder := httptest.NewRecorder()

	var json = []byte(`{"username":"Adphi", "password": "secret-password"}`)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(json))
	g.mux.ServeHTTP(recorder, req)

	// Make sure cookies is set properly
	hdr := recorder.Header()
	cookies, ok := hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	// Make sure db has one object
	var user testuser
	err := g.db.Where("username = ?", "Adphi").First(&user).Error
	if err != nil {
		t.Error("Fail to save object to database")
		return
	}

	// Make sure user is the same with current user from session
	logoutReq, _ := http.NewRequest("POST", "/auth/logout", nil)
	logoutReq.Header.Add("Cookie", cookies[0])
	currentUser, err := g.getCurrentUser(logoutReq)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(&user, currentUser) {
		t.Error("Get invalid current user from request")
	}

	// Logout
	recorder = httptest.NewRecorder()
	g.mux.ServeHTTP(recorder, logoutReq)

	// Make sure cookies is cleared after logout
	hdr = recorder.Header()
	cookies, ok = hdr["Set-Cookie"]
	if ok || len(cookies) == 1 {
		t.Fatal("Cookies should be cleared after logout")
	}

	// Test login
	loginReq, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(json))

	// Login
	recorder = httptest.NewRecorder()
	g.mux.ServeHTTP(recorder, loginReq)

	// Make sure cookies is set properly
	hdr = recorder.Header()
	cookies, ok = hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

}
