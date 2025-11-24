package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"uptime-monitor/models"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseMonitor handles monitoring different types of databases
type DatabaseMonitor struct {
	monitor *models.Monitor
}

// NewDatabaseMonitor creates a new database monitor instance
func NewDatabaseMonitor(monitor *models.Monitor) *DatabaseMonitor {
	return &DatabaseMonitor{monitor: monitor}
}

// Check performs the database health check
func (d *DatabaseMonitor) Check() (bool, string, int64, error) {
	start := time.Now()
	var isUp bool
	var message string
	var err error

	switch d.monitor.Type {
	case "mysql":
		isUp, message, err = d.checkMySQL()
	case "postgres":
		isUp, message, err = d.checkPostgres()
	case "mongodb":
		isUp, message, err = d.checkMongoDB()
	case "redis":
		isUp, message, err = d.checkRedis()
	default:
		return false, "Unsupported database type", 0, fmt.Errorf("unsupported database type: %s", d.monitor.Type)
	}

	responseTime := time.Since(start).Milliseconds()
	return isUp, message, responseTime, err
}

// checkMySQL checks MySQL database health
func (d *DatabaseMonitor) checkMySQL() (bool, string, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		d.monitor.DBUsername,
		d.monitor.DBPassword,
		d.monitor.DBHost,
		d.monitor.DBPort,
		d.monitor.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, "Failed to connect to MySQL", err
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return false, "Failed to ping MySQL", err
	}

	// Execute test query if provided
	if d.monitor.DBQuery != "" {
		var result string
		err := db.QueryRow(d.monitor.DBQuery).Scan(&result)
		if err != nil {
			return false, "Failed to execute test query", err
		}

		if result != d.monitor.DBExpectedValue {
			return false, "Test query result mismatch", nil
		}
	}

	return true, "MySQL is healthy", nil
}

// checkPostgres checks PostgreSQL database health
func (d *DatabaseMonitor) checkPostgres() (bool, string, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.monitor.DBHost,
		d.monitor.DBPort,
		d.monitor.DBUsername,
		d.monitor.DBPassword,
		d.monitor.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return false, "Failed to connect to PostgreSQL", err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false, "Failed to ping PostgreSQL", err
	}

	if d.monitor.DBQuery != "" {
		var result string
		err := db.QueryRow(d.monitor.DBQuery).Scan(&result)
		if err != nil {
			return false, "Failed to execute test query", err
		}

		if result != d.monitor.DBExpectedValue {
			return false, "Test query result mismatch", nil
		}
	}

	return true, "PostgreSQL is healthy", nil
}

// checkMongoDB checks MongoDB database health
func (d *DatabaseMonitor) checkMongoDB() (bool, string, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
		d.monitor.DBUsername,
		d.monitor.DBPassword,
		d.monitor.DBHost,
		d.monitor.DBPort,
		d.monitor.DBName,
	)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return false, "Failed to connect to MongoDB", err
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(context.Background(), nil); err != nil {
		return false, "Failed to ping MongoDB", err
	}

	if d.monitor.DBQuery != "" {
		// Execute test command
		var result bson.M
		cmd := bson.D{{Key: "eval", Value: d.monitor.DBQuery}}
		err := client.Database(d.monitor.DBName).RunCommand(context.Background(), cmd).Decode(&result)
		if err != nil {
			return false, "Failed to execute test command", err
		}

		if fmt.Sprintf("%v", result["retval"]) != d.monitor.DBExpectedValue {
			return false, "Test command result mismatch", nil
		}
	}

	return true, "MongoDB is healthy", nil
}

// checkRedis checks Redis database health
func (d *DatabaseMonitor) checkRedis() (bool, string, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", d.monitor.DBHost, d.monitor.DBPort),
		Password: d.monitor.DBPassword,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return false, "Failed to ping Redis", err
	}

	if d.monitor.DBQuery != "" {
		// Execute test command
		result, err := client.Do(ctx, d.monitor.DBQuery).Result()
		if err != nil {
			return false, "Failed to execute test command", err
		}

		if fmt.Sprintf("%v", result) != d.monitor.DBExpectedValue {
			return false, "Test command result mismatch", nil
		}
	}

	return true, "Redis is healthy", nil
}
