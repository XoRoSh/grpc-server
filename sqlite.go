package main

import (
	"fmt"
	"reflect"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Data struct {
	ID          string `gorm:"primaryKey"`
	Name        string `graphql:"name",description:"The name of the data entry"`
	Description string `graphql:"description",description:"The description of the data entry"`
}

type Car struct {
	VIN         string `gorm:"primaryKey"`
	Name        string `graphql:"name",description:"The name of the data entry"`
	Description string `graphql:"description",description:"The description of the data entry"`
}

type Structure struct {
	Description string
	Structure   interface{}
}

var e = []Structure{
	{Description: "Data structure", Structure: Data{}},
	{Description: "Car structure", Structure: Car{}},
}

var db *gorm.DB

func InitDB() (*gorm.DB, error) {
	var err error
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Perform schema migrations
	for _, entry := range e {
		// Use reflection to get the type of the Structure field
		structType := reflect.TypeOf(entry.Structure)

		// Ensure it's a pointer type
		if structType.Kind() == reflect.Struct {
			// Pass a pointer to the struct type to AutoMigrate
			ptr := reflect.New(structType).Interface()
			if err := db.AutoMigrate(ptr); err != nil {
				return nil, fmt.Errorf("failed to migrate %v: %w", structType.Name(), err)
			}
		}
	}

	// Insert sample data
	db.Create(&Data{ID: "1", Name: "Example", Description: "This is a sample data entry"})
	db.Create(&Car{VIN: "2", Name: "Another Example", Description: "This is another sample data entry"})

	return db, nil
}

func init() {
	_, err := InitDB()
	if err != nil {
		panic("failed to initialize database: " + err.Error())
	}
}
