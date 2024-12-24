// v3/config/config.go
package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

var (
	JwtSecretKey        string
	JwtRefreshSecretKey string
	AdminUsername       string
	AdminPassword       string
	DsnPostgres         string
	Environment         string
)

// LoadConfig Загружает переменные окружения из .env файла
func LoadConfig() {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Загружаем переменные окружения
	JwtSecretKey = os.Getenv("JWT_SECRET_KEY")
	JwtRefreshSecretKey = os.Getenv("JWT_REFRESH_SECRET_KEY")
	AdminUsername = os.Getenv("ADMIN_USERNAME")
	AdminPassword = os.Getenv("ADMIN_PASSWORD")
	DsnPostgres = os.Getenv("DSN_POSTGRES")
	Environment = os.Getenv("ENVIRONMENT") // development или production

	// Проверка на наличие обязательных переменных окружения
	requiredVars := []string{JwtSecretKey, JwtRefreshSecretKey, AdminUsername, AdminPassword, DsnPostgres, Environment}
	for _, v := range requiredVars {
		if v == "" {
			log.Fatalf("Missing required environment variable: %s", v)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	requiredFields  []string
	possibleFields  []string
	fieldsMutex     sync.RWMutex
	possibleFieldsM sync.RWMutex
)

// Устанавливаем обязательные поля для регистрации
func InitRequiredFields(fields []string) {
	fieldsMutex.Lock()
	defer fieldsMutex.Unlock()
	requiredFields = fields
}

// Получаем обязательные поля для регистрации
func GetRequiredFields() []string {
	fieldsMutex.RLock()
	defer fieldsMutex.RUnlock()
	return requiredFields
}

// Устанавливаем возможные поля для логина
func InitPossibleFields(fields []string) {
	possibleFieldsM.Lock()
	defer possibleFieldsM.Unlock()
	possibleFields = fields
}

// Получаем возможные поля для логина
func GetPossibleFields() []string {
	possibleFieldsM.RLock()
	defer possibleFieldsM.RUnlock()
	return possibleFields
}

// Проверяем, является ли поле возможным для логина
func IsPossibleField(field string) bool {
	possibleFieldsM.RLock()
	defer possibleFieldsM.RUnlock()
	for _, f := range possibleFields {
		if f == field {
			return true
		}
	}
	return false
}

//var (
//	requiredFields []string
//	fieldsMutex    sync.RWMutex
//)
//
//func InitRequiredFields(fields []string) {
//	fieldsMutex.Lock()
//	defer fieldsMutex.Unlock()
//	requiredFields = fields
//}
//
//func GetRequiredFields() []string {
//	fieldsMutex.RLock()
//	defer fieldsMutex.RUnlock()
//	return requiredFields
//}

//////

//type RegistrationConfig struct {
//	Method string // Возможные значения: "username", "email", "phone"
//}
//
//var registrationConfig atomic.Value
//
//func InitRegistrationConfig(method string) {
//	registrationConfig.Store(RegistrationConfig{Method: method})
//}
//
//func GetRegistrationMethod() string {
//	return registrationConfig.Load().(RegistrationConfig).Method
//}
