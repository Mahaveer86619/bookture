package db

import (
	"fmt"
	"log"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BooktureDB struct {
	DB *gorm.DB
}

var booktureDB = &BooktureDB{}

func InitBookture(isLocal ...bool) {
	dbHost := config.AppConfig.DB_HOST

	if len(isLocal) > 0 && isLocal[0] {
		dbHost = "localhost"
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost,
		config.AppConfig.DB_USER,
		config.AppConfig.DB_PASSWORD,
		config.AppConfig.DB_NAME,
		config.AppConfig.DB_PORT,
	)

	var err error
	booktureDB.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Database connection established successfully")
}

func GetBooktureDB() *BooktureDB {
	return booktureDB
}

func (bdb *BooktureDB) MigrateTables() {
	err := bdb.DB.AutoMigrate(
		&models.User{},
		&models.Library{},
		&models.Book{},
		&models.Volume{},
		&models.Chapter{},
		&models.Section{},
		&models.Scene{},
		&models.Character{},
		&models.CharacterVersion{},
		&models.CharacterRelationship{},
		&models.Summary{},
		&models.AIPrompt{},
		&models.AIGenerationJob{},
		&models.Progress{},
		&models.Bookmark{},
		&models.Rating{},
		&models.Annotation{},
		&models.Asset{},
		&models.EmbeddingTarget{},
		&models.Embedding{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed")
}