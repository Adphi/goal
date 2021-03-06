package goal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func queryPath(query []byte) string {
	return fmt.Sprint(testServer.URL, "/query/testuser/", url.QueryEscape(string(query)))
}

func createUsers() {
	// Prepare data
	names := []string{"Thomas", "Alan", "Jason", "Ben"}
	ages := []int{28, 30, 22, 40}
	for index, name := range names {
		user := &testuser{}
		user.Name = name
		user.Age = ages[index]
		g.db.Create(user)
	}
}

func TestSuccessQueryParamsFind(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()

	// Test success case

	item := &QueryItem{}
	item.Key = "name"
	item.Op = "="
	item.Val = "Thomas"

	params := g.NewQueryParams()
	params.Where = []*QueryItem{item}

	var results []testuser
	var user testuser
	params.Find(&user, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result. Got : ", results)
	}

	orItem := &QueryItem{}
	orItem.Key = "name"
	orItem.Op = "="
	orItem.Val = "Alan"
	item.Or = []*QueryItem{orItem}
	params.Where = []*QueryItem{item}
	params.Find(&user, &results)

	if results == nil || len(results) != 2 {
		t.Error("Error: query should return 2 result. Got : ", results)
	}

	andItem := &QueryItem{}
	andItem.Key = "age"
	andItem.Op = ">"
	andItem.Val = "29"
	params.Where = []*QueryItem{item, andItem}
	params.Find(&user, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result. Got : ", results)
	}

}

func TestSuccessQueryBuilderFind(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()

	var results []testuser
	var user testuser

	q := g.NewQuery().Where("name").Equals("Thomas")
	q.Find(&user, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result. Got : ", results)
	}

	q = q.Or("name").Equals("Alan")
	q.Find(&user, &results)

	if results == nil || len(results) != 2 {
		t.Error("Error: query should return 2 result. Got : ", results)
	}

	q.And("age").Sup(29).Find(&user, &results)
	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result. Got : ", results)
	}
}

func TestInvalidQueryParamsFind(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()

	// Test success case

	item := &QueryItem{}
	item.Key = "name"
	item.Op = "hello"
	item.Val = "Thomas"

	params := g.NewQueryParams()
	params.Where = []*QueryItem{item}

	var results []testuser
	var user testuser
	err := params.Find(&user, &results)
	if err == nil {
		t.Error("Error: query operator should be invalid")
	}

	item = &QueryItem{}
	item.Key = "hello"
	item.Op = "="
	item.Val = "Thomas"
	params.Where = []*QueryItem{item}

	err = params.Find(&user, &results)
	if err == nil {
		t.Error("Error: query column should be invalid")
	}
}

func TestQueryViaAPI(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()
	item := &QueryItem{}
	item.Key = "name"
	item.Op = "="
	item.Val = "Thomas"

	params := g.NewQueryParams()
	params.Where = []*QueryItem{item}

	query, _ := json.Marshal(params)

	req, _ := http.NewRequest("GET", queryPath(query), nil)
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	if res.StatusCode != 200 {
		fmt.Println(res.StatusCode)
		t.Error("Request Failed")
		return
	}

	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var results []testuser
	json.Unmarshal(content, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result. Got : ", results)
	}
}
