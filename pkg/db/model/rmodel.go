package model

import (
	"fmt"
	"reflect"

	r "github.com/dancannon/gorethink"
	"github.com/dancannon/gorethink/encoding"
	"github.com/materials-commons/mcstore/pkg/app"
	"github.com/materials-commons/mcstore/pkg/db"
)

// rModel holds the schema definition and the table for the schema.
type rModel struct {
	schema interface{}
	table  string
}

// rQuery holds the model and database references, such as the query to run
// and the database session.
type rQuery struct {
	*rModel
	Rql     r.Term
	Session *r.Session
}

// ByID retrieves an entry by its id field.
func (q *rQuery) ByID(id string, obj interface{}) error {
	err := GetItem(id, q.table, q.Session, obj)
	return err
}

// Q constructs a rQuery and fills in its Session by calling db.RSession().
func (m *rModel) Q() *rQuery {
	session, err := db.RSession()
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database: %s", err))
	}
	return m.Qs(session)
}

// Qs constructs a query and accepts a database Session to use.
func (m *rModel) Qs(session *r.Session) *rQuery {
	return &rQuery{
		rModel:  m,
		Session: session,
		Rql:     r.Table(m.table),
	}
}

// Row returns a single item. It takes an arbitrary query.
func (q *rQuery) Row(query r.Term, obj interface{}) error {
	err := GetRow(query, q.Session, obj)
	return err
}

// Table returns the Rql for the table. It abstracts away having to know the particular
// table for a given model.
func (m *rModel) Table() r.Term {
	return r.Table(m.table)
}

// T is a shortcut for Table.
func (m *rModel) T() r.Term {
	return r.Table(m.table)
}

// Rows returns a list of items from the database. It takes an arbitrary query.
func (q *rQuery) Rows(query r.Term, results interface{}) error {
	elementType := reflect.TypeOf(q.schema)
	resultsValue := reflect.ValueOf(results)
	if resultsValue.Kind() != reflect.Ptr || (resultsValue.Elem().Kind() != reflect.Slice && resultsValue.Elem().Kind() != reflect.Interface) {
		return fmt.Errorf("bad type for results")
	}

	sliceValue := resultsValue.Elem()

	if resultsValue.Elem().Kind() == reflect.Interface {
		sliceValue = sliceValue.Elem().Slice(0, sliceValue.Cap())
	} else {
		sliceValue = sliceValue.Slice(0, sliceValue.Cap())
	}

	rows, err := query.Run(q.Session)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 0
	var result = reflect.New(elementType)
	for rows.Next(result.Interface()) {
		if sliceValue.Len() == i {
			sliceValue = reflect.Append(sliceValue, result.Elem())
			sliceValue = sliceValue.Slice(0, sliceValue.Cap())
		} else {
			sliceValue.Index(i).Set(result.Elem())
		}
		i++
		result = reflect.New(elementType)
	}

	resultsValue.Elem().Set(sliceValue.Slice(0, i))
	return nil
}

// Update updates an existing database model entry.
func (q *rQuery) Update(id string, what interface{}) error {
	var (
		dv  interface{}
		err error
	)
	v := reflect.ValueOf(what)
	if v.Kind() == reflect.Struct || v.Kind() == reflect.Struct {
		dv, err = encoding.Encode(what)
		if err != nil {
			return app.ErrInvalid
		}
	} else {
		dv = what
	}
	rv, err := q.T().Get(id).Update(dv).RunWrite(q.Session)
	switch {
	case err != nil:
		return err
	case rv.Errors != 0:
		return app.ErrNotFound
	default:
		return nil
	}
}

// Insert a new entry into the database using the specified table. This is
// available so that models can be used to update dependent tables without
// having to create a model for them. For example, a join table doesn't
// really need a model.
func (q *rQuery) InsertRaw(table string, what interface{}, dest interface{}) error {
	returnValue := false
	dv := reflect.ValueOf(dest)
	if dv.Kind() == reflect.Ptr {
		returnValue = true
	} else if dv.Kind() != reflect.Invalid {
		return app.ErrInvalid
	}

	opts := r.InsertOpts{
		ReturnChanges: returnValue,
		Durability:    "hard",
	}

	rv, err := r.Table(table).Insert(what, opts).RunWrite(q.Session)
	switch {
	case err != nil:
		return err
	case rv.Errors != 0:
		return app.ErrCreate
	case rv.Inserted == 0:
		return app.ErrCreate
	case returnValue:
		if len(rv.Changes) == 0 {
			return app.ErrCreate
		}
		err := encoding.Decode(dest, rv.Changes[0].NewValue)
		return err
	default:
		return nil
	}
}

// Insert inserts a new model entry into the database
func (q *rQuery) Insert(what interface{}, dest interface{}) error {
	return q.InsertRaw(q.table, what, dest)
}

// Delete deletes an existing database model entry.
func (q *rQuery) Delete(id string) error {
	rv, err := q.T().Get(id).Delete().RunWrite(q.Session)
	switch {
	case err != nil:
		return err
	case rv.Errors != 0:
		return app.ErrNotFound
	case rv.Deleted == 0:
		return app.ErrNotFound
	default:
		return nil
	}
}

// GetItem retrieves an item by id in the given table.
func GetItem(id, table string, session *r.Session, obj interface{}) error {
	result, err := r.Table(table).Get(id).Run(session)
	switch {
	case err != nil:
		return err
	case result.IsNil():
		return app.ErrNotFound
	default:
		err := result.One(obj)
		return err
	}
}

// GetRow runs a query and returns a single item.
func GetRow(query r.Term, session *r.Session, obj interface{}) error {
	result, err := query.Run(session)
	switch {
	case err != nil:
		return err
	case result.IsNil():
		return app.ErrNotFound
	default:
		err := result.One(obj)
		return err
	}
}

// GetRows runs a query an returns a list of results.
func GetRows(query r.Term, session *r.Session, results interface{}) error {
	resultsValue := reflect.ValueOf(results)
	if resultsValue.Kind() != reflect.Ptr || (resultsValue.Elem().Kind() != reflect.Slice && resultsValue.Elem().Kind() != reflect.Interface) {
		return fmt.Errorf("bad type for results")
	}

	sliceValue := resultsValue.Elem()

	if resultsValue.Elem().Kind() == reflect.Interface {
		sliceValue = sliceValue.Elem().Slice(0, sliceValue.Cap())
	} else {
		sliceValue = sliceValue.Slice(0, sliceValue.Cap())
	}
	elementType := sliceValue.Type().Elem()
	rows, err := query.Run(session)
	if err != nil {
		return err
	}

	defer rows.Close()

	i := 0
	var result = reflect.New(elementType)
	for rows.Next(result.Interface()) {
		if sliceValue.Len() == i {
			sliceValue = reflect.Append(sliceValue, result.Elem())
			sliceValue = sliceValue.Slice(0, sliceValue.Cap())
		} else {
			sliceValue.Index(i).Set(result.Elem())
		}
		i++
		result = reflect.New(elementType)
	}

	resultsValue.Elem().Set(sliceValue.Slice(0, i))
	return nil
}

// Delete deletes an item by id in the given table.
func Delete(table, id string, session *r.Session) error {
	_, err := r.Table(table).Get(id).Delete().RunWrite(session)
	return err
}
