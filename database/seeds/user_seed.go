package seeds

import (
	"doc-api/internal/core/domain/model"
	"doc-api/utils"
	"log"

	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB) {
	bytes, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("%s: %v", err.Error(), err)
	}

	admin := model.User{
		Name:     "super admin",
		Email:    "superadmin@mail.com",
		Password: bytes,
	}

	if err := db.FirstOrCreate(&admin, model.User{Email: "superadmin@mail.com"}).Error; err != nil {
		log.Fatalf("%s: %v", err.Error(), err)
	} else {
		log.Printf("Admin %s created", admin.Name)
	}
}
