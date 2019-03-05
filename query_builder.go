package goal

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
)

type query struct {
	db    *gorm.DB
	w     []*QueryItem
	l     int64
	s     int64
	i     []string
	o     map[string]bool
	Error error
}

func (g *Goal) NewQuery() *query {
	return &query{db: g.db}
}

func (q *query) Where(key string) *query {
	q.w = append(q.w, &QueryItem{Key: key})
	return q
}

func (q *query) And(key string) *query {
	if q.Error = q.validate(); q.Error != nil {
		return q
	}
	q.w = append(q.w, &QueryItem{Key: key})
	return q
}

func (q *query) Or(key string) *query {
	if q.Error = q.validate(); q.Error != nil {
		return q
	}
	i := lastItem(q.w)
	if i.Val == nil {
		q.Error = ErrNoValue
		return q
	}
	i.Or = append(i.Or, &QueryItem{Key: key})
	return q
}

func (q *query) Limit(i int64) *query {
	q.l = i
	return q
}

func (q *query) Skip(i int64) *query {
	q.s = i
	return q
}

func (q *query) Order(key string, order Order, reorder bool) *query {
	s := fmt.Sprint(key, " ", order)
	if q.o == nil {
		q.o = map[string]bool{}
	}
	q.o[s] = reorder
	return q
}

func (q *query) Include(relation ...string) *query {
	q.i = relation
	return q
}

func (q *query) op(op Op, val interface{}) *query {
	if q.Error = q.validate(); q.Error != nil {
		return q
	}
	i := lastItem(q.w)
	if i.Or != nil && len(i.Or) > 0 {
		if lastItem(i.Or).Key == "" {
			q.Error = ErrNoKey
			return q
		}
		o := lastItem(i.Or)
		o.Op = op
		o.Val = val
		return q
	}

	i.Op = op
	i.Val = val
	return q
}

func (q *query) Equals(val interface{}) *query {
	return q.op(Equal, val)
}
func (q *query) Sup(val interface{}) *query {
	return q.op(Sup, val)
}
func (q *query) SupEq(val interface{}) *query {
	return q.op(SupEq, val)
}
func (q *query) Inf(val interface{}) *query {
	return q.op(Inf, val)
}
func (q *query) InfEq(val interface{}) *query {
	return q.op(InfEq, val)
}
func (q *query) NotEq(val interface{}) *query {
	return q.op(NotEq, val)
}
func (q *query) In(val ...interface{}) *query {
	return q.op(In, val)
}
func (q *query) Like(val interface{}) *query {
	return q.op(Like, val)
}

func (q *query) Find(resource interface{}, results interface{}) error {
	p := queryParams{
		db:      q.db,
		Where:   q.w,
		Order:   q.o,
		Limit:   q.l,
		Skip:    q.s,
		Include: q.i,
	}

	return p.Find(resource, results)
}

func (q *query) validate() error {
	if len(q.w) == 0 || lastItem(q.w).Key == "" {
		return ErrNoKey
	}
	return q.Error
}

func lastItem(items []*QueryItem) *QueryItem {
	return items[len(items)-1]
}

var (
	ErrNoKey   = errors.New("no query key selected")
	ErrNoValue = errors.New("no query value selected")
)
