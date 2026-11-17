package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Use the local development database by default. Set DB_HOST and the other
	// DB_* variables in hosted environments; TiDB Cloud requires TLS.
	host := envOrDefault("DB_HOST", "127.0.0.1")
	port := envOrDefault("DB_PORT", "3306")
	dbName := envOrDefault("DB_NAME", "school_management")
	user := envOrDefault("DB_USER", "root")
	password := os.Getenv("DB_PASSWORD")

	dsnConfig := mysqlDriver.Config{
		User:      user,
		Passwd:    password,
		Net:       "tcp",
		Addr:      net.JoinHostPort(host, port),
		DBName:    dbName,
		ParseTime: true,
		Loc:       time.Local,
	}
	if os.Getenv("DB_TLS") == "true" {
		dsnConfig.TLSConfig = "true"
	}
	dsn := dsnConfig.FormatDSN()

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	DB = database

	fmt.Println("Database berhasil terhubung!")
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
