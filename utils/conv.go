package utils

import (
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	start := time.Now()

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("bcrypt compare")

	return err == nil
}
