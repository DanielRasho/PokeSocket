package postgres_cli

import (
	"fmt"
	"strconv"
	"strings"
)

// QueryBuilder helps to build Postgres Queries dynamically and protecting from sql injection.
type QueryBuilder struct {
	b          strings.Builder
	params     []any
	onBlock    bool
	numClauses int
}

func (q *QueryBuilder) Query(s string) *QueryBuilder {
	q.b.WriteString(" " + s + " ")
	return q
}

// Resets the QueryBuilder completely. Useful for recycling the queryBuilder
// instead of creating a new one
func (q *QueryBuilder) Flush() {
	q.b.Reset()
	q.params = nil
	q.onBlock = false
	q.numClauses = 0
}

// Adds a parameter to query.
func (q *QueryBuilder) Param(val any) *QueryBuilder {
	q.b.WriteString(fmt.Sprintf("$%d ", len(q.params)+1))
	q.params = append(q.params, val)
	return q
}

// Returns the final query sentence and the params that it requests.
func (q *QueryBuilder) Get() (string, []any) {
	return q.b.String(), q.params
}

func (q *QueryBuilder) StartBlock() {
	q.onBlock = true
}

func (q *QueryBuilder) EndBlock() {
	q.onBlock = false
	q.numClauses = 0
}

func (q *QueryBuilder) Clause(prefix, query string, values ...any) {
	// Generate args
	args := make([]any, 0, len(values))
	for _, v := range values {
		args = append(args, "$"+strconv.Itoa(len(q.params)+1))
		q.params = append(q.params, v)
	}

	// Add boolean operator if no condition has ben added
	if q.numClauses > 0 {
		q.b.WriteString(" " + prefix + " ")
	}

	// Add parameters
	q.b.WriteString(fmt.Sprintf(query, args...))
	q.numClauses++
}
