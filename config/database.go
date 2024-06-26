package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

func Connect() {
    var err error
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
        os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_PASSWORD"),
    )

    DB, err = gorm.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }
    log.Println("Database connection successful")
}

func GetDB() *gorm.DB {
    return DB
}
