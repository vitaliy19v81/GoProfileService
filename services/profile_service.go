// /home/vtoroy/GolandProjects/profile_service/services/profile_service.go
package services

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type Profile struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	FirstName      string                 `json:"first_name"`
	LastName       string                 `json:"last_name"`
	MiddleName     sql.NullString         `json:"middle_name,omitempty"`
	Phone          sql.NullString         `json:"phone,omitempty"`
	Address        sql.NullString         `json:"address,omitempty"`
	Birthday       sql.NullTime           `json:"birthday,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

///////////

func CreateProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// (s *ProfileService) c *gin.Context
		var profile Profile
		if err := c.ShouldBindJSON(&profile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные профиля"})
			return
		}

		// Извлекаем user_id из контекста
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		// Преобразуем userID в строку, если это необходимо
		profile.UserID = userID.(string)

		if db == nil {
			log.Println("Database connection is not initialized")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		query := `
		INSERT INTO profiles (user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

		additionalDataJSON, err := json.Marshal(profile.AdditionalData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сериализации данных"})
			return
		}

		err = db.QueryRow(query,
			profile.UserID, profile.FirstName, profile.LastName, profile.MiddleName,
			profile.Phone, profile.Address, profile.Birthday, additionalDataJSON).
			Scan(&profile.ID, &profile.CreatedAt, &profile.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания профиля"})
			return
		}
		c.JSON(http.StatusCreated, profile)
	}
}

func GetProfileByToken(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			log.Println(token)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Токен не указан"})
			return
		}

		var profile Profile
		query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
                  FROM profiles WHERE token = $1`
		err := db.QueryRow(query, token).Scan(
			&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.MiddleName,
			&profile.Phone, &profile.Address, &profile.Birthday, &profile.AdditionalData,
			&profile.CreatedAt, &profile.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Профиль не найден"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения профиля"})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

func GetProfileByUserID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		var profile Profile
		query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
                  FROM profiles WHERE user_id = $1`
		err := db.QueryRow(query, userID).Scan(
			&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.MiddleName,
			&profile.Phone, &profile.Address, &profile.Birthday, &profile.AdditionalData,
			&profile.CreatedAt, &profile.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Профиль не найден"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения профиля"})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
}

func GetProfile(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var profile Profile

		query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
			  FROM profiles WHERE id = $1`
		err := db.QueryRow(query, id).Scan(
			&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.MiddleName,
			&profile.Phone, &profile.Address, &profile.Birthday, &profile.AdditionalData,
			&profile.CreatedAt, &profile.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Профиль не найден"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения профиля"})
			return
		}
		c.JSON(http.StatusOK, profile)
	}
} // c *gin.Context

//type ProfileService struct {
//	ps.UnimplementedProfileServiceServer
//	db *sql.DB
//}
//
//func NewProfileService(db *sql.DB) *ProfileService {
//	if db == nil {
//		log.Fatal("Database connection is nil")
//	}
//	return &ProfileService{db: db}
//}

////////

//func (s *ProfileService) GetProfile() {
//
//		id := c.Param("id")
//		var profile Profile
//
//		query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
//			  FROM profiles WHERE id = $1`
//		err := s.db.QueryRow(query, id).Scan(
//			&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.MiddleName,
//			&profile.Phone, &profile.Address, &profile.Birthday, &profile.AdditionalData,
//			&profile.CreatedAt, &profile.UpdatedAt)
//		if err == sql.ErrNoRows {
//			c.JSON(http.StatusNotFound, gin.H{"error": "Профиль не найден"})
//			return
//		} else if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения профиля"})
//			return
//		}
//		c.JSON(http.StatusOK, profile)
//	} // c *gin.Context
