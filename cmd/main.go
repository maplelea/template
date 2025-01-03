package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var ctx = context.Background()

func initViper() {
	// 设置配置文件的名称（不带扩展名）
	viper.SetConfigName("config")

	// 设置配置文件所在的路径
	viper.AddConfigPath("./config")

	// 支持的配置文件类型（yaml、json、toml 等）
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("fatal error config file: %s", err)
	}
}

func connectMySQL() *gorm.DB {
	// 从配置文件中读取 MySQL DSN
	dsn := viper.GetString("database.dsn")

	// 打开 MySQL 连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}

	// 测试连接
	if err := db.Exec("SELECT 1").Error; err != nil {
		log.Fatalf("failed to ping MySQL: %v", err)
	}

	fmt.Println("Successfully connected to MySQL!")
	return db
}

func connectRedis() *redis.Client {
	// 从配置文件中读取 Redis 配置
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	redisDB := viper.GetInt("redis.db")

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	fmt.Println("Successfully connected to Redis!")
	return rdb
}

func connectRabbitMQ() *amqp.Connection {
	// 从配置文件中读取 RabbitMQ URL
	rabbitMQURL := viper.GetString("rabbitmq.url")

	// 连接到 RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	fmt.Println("Successfully connected to RabbitMQ!")
	return conn
}

func main() {
	// 初始化 Viper 并加载配置文件
	initViper()

	// 连接 MySQL
	gormDB := connectMySQL()

	// 获取底层的 *sql.DB 并延迟关闭
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get underlying *sql.DB: %v", err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {
			log.Fatalf("failed to close DB: %v", err)
		}
	}(sqlDB)

	// 连接 Redis
	rdb := connectRedis()
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			log.Fatalf("failed to close Redis: %v", err)
		}
	}(rdb)

	// 连接 RabbitMQ
	rabbitMQConn := connectRabbitMQ()
	defer func(rabbitMQConn *amqp.Connection) {
		err := rabbitMQConn.Close()
		if err != nil {
			log.Fatalf("failed to close RabbitMQ: %v", err)
		}
	}(rabbitMQConn)
}
