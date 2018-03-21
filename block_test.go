package db2sql

import (
	"fmt"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// TestGetInputAddress will test if we have a right hex to bitcoin address translator
func TestInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO block").
		WithArgs(2, 3).
		WillReturnError(fmt.Errorf("some error"))
}

func TestConvert(t *testing.T) {

}
