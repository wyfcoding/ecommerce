package main

import (
	"ecommerce/internal/cms/biz"
	"ecommerce/internal/cms/data"
	"ecommerce/internal/cms/service"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config holds the application configuration.
type Config struct {
	Service  ServiceConfig  `toml:"service"`
	Database DatabaseConfig `toml:"database"`
}

type ServiceConfig struct {
	Port string `toml:"port"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`	DBName   string `toml:"dbname"`
}

func main() {
	// ======== 1. Initialize Dependencies (e.g., Config, Logger, DB) ======== 

	// Load configuration from TOML file
	var conf Config
	if _, err := toml.DecodeFile("../../configs/cms.toml", &conf); err != nil {
		log.Fatalf("failed to load config file: %v", err)
	}

	// Construct MySQL DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.DBName,
	)

	// Connect to MySQL database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&data.ContentPage{}, &data.ContentBlock{})
	if err != nil {
		log.Fatalf("failed to migrate schema: %v", err)
	}

	log.Println("Successfully connected to database and migrated schema.")

	// ======== 2. Wire up the application layers (Dependency Injection) ======== 

	dataRepo, cleanup, err := data.NewData(db)
	if err != nil {
		log.Fatalf("failed to create data layer: %v", err)
	}
	defer cleanup()

	cmsRepo := data.NewCmsRepo(dataRepo)
	cmsUsecase := biz.NewCmsUsecase(cmsRepo)
	cmsService := service.NewCmsService(cmsUsecase)

	log.Println("Application layers wired successfully.")

	// ======== 3. Start the Server (e.g., HTTP, gRPC) ======== 

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fw, _ := fmt.Fprint(w, "OK")
	})

	portStr := fmt.Sprintf(":%s", conf.Service.Port)
	log.Printf("Starting HTTP server on port %s", portStr)

	if err := http.ListenAndServe(portStr, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	_ = cmsService
}
