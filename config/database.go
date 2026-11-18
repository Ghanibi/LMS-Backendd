package config

import (
	"fmt"
	"net"
	"os"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	// Use the local development database by default. Set DB_HOST and the other
	// DB_* variables in hosted environments; TiDB Cloud requires TLS.
	host := envOrDefault("DB_HOST", "127.0.0.1")
	port := envOrDefault("DB_PORT", "3306")
	dbName := envOrDefault("DB_NAME", "school_management")
	user := envOrDefault("DB_USER", "root")
	password := os.Getenv("DB_PASSWORD")

	// NewConfig applies driver defaults, including allowing MySQL native
	// password authentication used by TiDB Cloud.
	dsnConfig := mysqlDriver.NewConfig()
	dsnConfig.User = user
	dsnConfig.Passwd = password
	dsnConfig.Net = "tcp"
	dsnConfig.Addr = net.JoinHostPort(host, port)
	dsnConfig.DBName = dbName
	dsnConfig.ParseTime = true
	dsnConfig.Loc = time.Local
	dsnConfig.Timeout = 10 * time.Second
	dsnConfig.ReadTimeout = 10 * time.Second
	dsnConfig.WriteTimeout = 10 * time.Second
	if os.Getenv("DB_TLS") == "true" {
		dsnConfig.TLSConfig = "true"
	}
	dsn := dsnConfig.FormatDSN()

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("gagal terhubung ke database: %w", err)
	}

	DB = database

	fmt.Println("Database berhasil terhubung!")
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
