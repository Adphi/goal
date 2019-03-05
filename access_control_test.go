package goal

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Satisfy Roler interface
func (user *testuser) Roles() []string {
	ownRole := fmt.Sprintf("testuser:%v", user.ID)
	roles := []string{ownRole}

	return roles
}

func (art *article) Get(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return g.read(reflect.TypeOf(art), request)
}

func (art *article) Post(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return g.create(reflect.TypeOf(art), request)
}

func (art *article) Query(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return g.handleQuery(reflect.TypeOf(art), request)
}

func TestCanRead(t *testing.T) {
	setup()
	defer tearDown()

	// Create article with author
	author := &testuser{}
	author.Username = "secret"
	g.db.Create(author)

	art := &article{}
	art.Author = author
	art.Permission = Permission{
		Read:  `["admin", "ceo"]`,
		Write: `["admin", "ceo"]`,
	}
	art.Title = "Top Secret"

	err := g.db.Create(art).Error
	if err != nil {
		fmt.Println("error Create article ", err)
	}

	res := httptest.NewRecorder()

	var json = []byte(`{"username":"Adphi", "password": "something-secret"}`)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(json))

	g.mux.ServeHTTP(res, req)

	// Make sure cookies is set properly
	hdr := res.Header()
	cookies, ok := hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	artURL := fmt.Sprint(testServer.URL, "/article/", art.ID)

	// Make sure user is the same with current user from session
	nextReq, _ := http.NewRequest("GET", artURL, nil)
	nextReq.Header.Add("Cookie", cookies[0])

	// Get response
	client := &http.Client{}
	resp, err := client.Do(nextReq)
	resp.Body.Close()

	if resp.StatusCode != 403 || err != nil {
		t.Error("Request should be unauthorized because Adphi doesn't have admin role")
	}
}
