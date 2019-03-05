// Define data structure for a query request
// {
//   "where":[{"key": "name", "op": "=", "val": "Thomas"}],
//   "order": [{"key": "name", "val": "asc"}]
//   "limit": 1
// }

package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
)

var ops map[Op]bool

type Op string

type Order string

const (
	Equal Op = "="
	Sup   Op = ">"
	SupEq Op = ">="
	Inf   Op = "<"
	InfEq Op = "<="
	NotEq Op = "<>"
	In    Op = "in"
	Like  Op = "like"

	Asc  Order = "ASC"
	Desc Order = "DESC"
)

func allowedOps() map[Op]bool {
	if ops == nil {
		ops = map[Op]bool{
			Equal: true,
			Sup:   true,
			SupEq: true,
			Inf:   true,
			InfEq: true,
			NotEq: true,
			In:    true,
			Like:  true,
		}
	}
	return ops
}

// QueryItem defines most basic element of a query.
// For example: name = Thomas
type QueryItem struct {
	Key string       `json:"key"`
	Op  Op           `json:"op"`
	Val interface{}  `json:"val"`
	Or  []*QueryItem `json:"or"`
}

// queryParams defines structure of a query. Where clause
// may include multiple QueryItem and connect by "AND" operator
type queryParams struct {
	db      *gorm.DB        `json:"-"`
	Where   []*QueryItem    `json:"where"`
	Limit   int64           `json:"limit"`
	Skip    int64           `json:"skip"`
	Order   map[string]bool `json:"order"`
	Include []string        `json:"include"`
}

func (g *Goal) NewQueryParams() *queryParams {
	return &queryParams{db: g.db}
}

// Find constructs the query, return error immediately if query is invalid,
// and query database if everything is valid
func (params *queryParams) Find(resource interface{}, results interface{}) error {
	scope := params.db.NewScope(resource)

	qryDB := params.db.New()

	// Parse where clause
	if params.Where != nil {
		for _, item := range params.Where {
			query, err := item.getQuery(scope)

			// Return immediately if query is invalid
			if err != nil {
				return err
			}

			qryDB = qryDB.Where(query, item.Val)

			if item.Or != nil {
				for _, orItem := range item.Or {
					query, err = orItem.getQuery(scope)

					// Return immediately if query is invalid
					if err != nil {
						return err
					}

					qryDB = qryDB.Or(query, orItem.Val)
				}
			}
		}
	}

	if params.Limit != 0 {
		qryDB = qryDB.Limit(params.Limit)
	}

	if params.Skip != 0 {
		qryDB = qryDB.Offset(params.Skip)
	}

	if params.Order != nil {
		for name, order := range params.Order {
			name = strings.Title(name)
			if !scope.HasColumn(name) {
				errorMsg := fmt.Sprintf("Column %s does not exist", name)
				return errors.New(errorMsg)
			}

			qryDB = qryDB.Order(name, order)
		}
	}

	if params.Include != nil {
		for _, name := range params.Include {
			qryDB = qryDB.Preload(strings.Title(name))
		}
	}

	// query the database
	qryDB.Find(results)

	return nil
}

// HandleQuery retrieves results filtered by request parameters
func (g *Goal) handleQuery(rType reflect.Type, request *http.Request) (int, interface{}, error) {
	vars := mux.Vars(request)

	// Retrieve query parameter
	query, err := url.QueryUnescape(vars["query"])

	if err != nil {
		return 500, nil, err
	}

	params := g.NewQueryParams()
	err = json.Unmarshal([]byte(query), &params)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	resource := newObjectWithType(rType)
	results := dynamicSlice(resource)

	err = params.Find(resource, results)
	if err != nil {
		return 500, nil, err
	}

	// Check permission for each item, remove item which doesn't have permission
	var filtered []interface{}

	switch reflect.TypeOf(results).Elem().Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(results).Elem()

		for i := 0; i < s.Len(); i++ {
			item := s.Index(i).Interface()
			err = g.CanPerform(item, request, true)

			// Only add to the filtered slice if no permission error
			if err == nil {
				filtered = append(filtered, item)
			}
		}
	default:
		panic("results should be a slice")
	}

	return 200, filtered, nil
}

func (item *QueryItem) getQuery(scope *gorm.Scope) (string, error) {
	_, exists := allowedOps()[item.Op]
	if !exists {
		str := fmt.Sprintf("Invalid SQL operator: %s", item.Op)
		return "", errors.New(str)
	}

	if !scope.HasColumn(item.Key) {
		str := fmt.Sprintf("Column does not exist: %s", item.Key)
		return "", errors.New(str)
	}

	var query string

	if item.Op == "in" {
		query = fmt.Sprintf("%s %s (?)", item.Key, item.Op)
	} else {
		query = fmt.Sprintf("%s %s ?", item.Key, item.Op)
	}

	return query, nil
}
