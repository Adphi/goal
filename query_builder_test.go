package goal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	q := NewQuery()
	q.And("key")
	assert.Error(t, q.Error)
	q = NewQuery()

	q.Where("myKey")
	assert.NoError(t, q.Error)
	assert.Equal(t, 1, len(q.w))
	assert.Equal(t, "myKey", q.w[0].Key)

	q.Equals("some value")
	assert.NoError(t, q.Error)
	assert.Equal(t, 1, len(q.w))
	assert.Equal(t, "myKey", q.w[0].Key)
	assert.Equal(t, Equal, q.w[0].Op)
	assert.Equal(t, "some value", q.w[0].Val)

	q.And("other")
	assert.NoError(t, q.Error)
	assert.Equal(t, 2, len(q.w))
	assert.Equal(t, "other", q.w[1].Key)

	q.Like("other value")
	assert.NoError(t, q.Error)
	assert.Equal(t, 2, len(q.w))
	assert.Equal(t, Like, q.w[1].Op)
	assert.Equal(t, "other value", q.w[1].Val)

	q.Or("also")
	assert.NoError(t, q.Error)
	assert.Equal(t, 2, len(q.w))
	assert.Equal(t, 1, len(q.w[1].Or))
	assert.Equal(t, "also", q.w[1].Or[0].Key)

	q.In("first", "second")
	assert.NoError(t, q.Error)
	assert.Equal(t, 2, len(q.w))
	assert.Equal(t, 1, len(q.w[1].Or))
	assert.Equal(t, In, q.w[1].Or[0].Op)
	assert.Equal(t, []interface{}{"first", "second"}, q.w[1].Or[0].Val)

	q.And("finally")
	assert.NoError(t, q.Error)
	assert.Equal(t, 3, len(q.w))
	assert.Equal(t, "finally", q.w[2].Key)

	q.Sup(3)
	assert.NoError(t, q.Error)
	assert.Equal(t, 3, len(q.w))
	assert.Equal(t, 3, q.w[2].Val)
	assert.Equal(t, Sup, q.w[2].Op)
}

func TestFullQuery(t *testing.T) {
	q := NewQuery().
		Where("id").
		Equals(1).
		Or("user").
		Like("admin").
		And("active").
		NotEq(false).
		Order("id", Asc, true).
		Limit(10).
		Skip(20)
	expected := &query{
		w: []*QueryItem{
			{
				Key: "id",
				Op:  Equal,
				Val: 1,
				Or: []*QueryItem{{
					Key: "user",
					Op:  Like,
					Val: "admin",
				}},
			}, {
				Key: "active",
				Op:  NotEq,
				Val: false,
			},
		},
		o: map[string]bool{"id ASC": true},
		l: 10,
		s: 20,
	}
	assert.NoError(t, q.Error)
	assert.Equal(t, expected, q)
}

func TestQueryError(t *testing.T) {
	q := NewQuery().Where("key").Or("other_key").Equals(false)
	assert.Error(t, q.Error)
}
