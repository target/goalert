package gqlauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	q, err := NewQuery(`query ($id: ID!) { user(id: $id) { id, name } }`)
	require.NoError(t, err)

	isSub, err := q.IsSubset(`{ user { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

	isSub, err = q.IsSubset(`query { user(id: "afs") { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

	isSub, err = q.IsSubset(`query { user { id, name, email } }`)
	require.NoError(t, err)
	assert.False(t, isSub)
}
func TestQueryObj(t *testing.T) {

	q, err := NewQuery(`query ($in: DebugMessagesInput) { debugMessages(input: $in) { id } }`)
	require.NoError(t, err)

	isSub, err := q.IsSubset(`{ debugMessages { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

	isSub, err = q.IsSubset(`{ debugMessages(input:{first: 3}) { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

}
func TestQueryObjLim(t *testing.T) {

	q, err := NewQuery(`query ($first: Int) { debugMessages(input: {first: $first}) { id } }`)
	require.NoError(t, err)

	isSub, err := q.IsSubset(`{ debugMessages { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

	isSub, err = q.IsSubset(`{ debugMessages(input:{first: 3}) { id } }`)
	require.NoError(t, err)
	assert.True(t, isSub)

	isSub, err = q.IsSubset(`{ debugMessages(input:{createdBefore: "asdf"}) { id } }`)
	require.NoError(t, err)
	assert.False(t, isSub)

}
