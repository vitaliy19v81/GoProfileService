// /home/vtoroy/GolandProjects/profile_service/cmd/api/main.go
package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"profile_service/config"
	"profile_service/database"
	"profile_service/internal/services/annotation_proto"
	authpb "profile_service/internal/services/auth_proto"
	"profile_service/middleware"
	ps "profile_service/profile_service"
	rProfile "profile_service/services"

	"github.com/gin-gonic/gin"
)

//////////////////////////////////////////////////////////////////////////

//go get -u github.com/swaggo/swag/cmd/swag
//go get -u github.com/swaggo/gin-swagger
//go get -u github.com/swaggo/files

//export PATH=$PATH:$(go env GOPATH)/bin
//source ~/.bashrc
//echo $PATH
//swag init -g cmd/api/main.go -o /docs

//////////////////////////////////////////////////////////////////////////
// Генерация документации
// protoc \
//  -I /home/vtoroy/GolandProjects/profile_service/proto \
//  -I /home/vtoroy/GolandProjects/profile_service/api-common-protos \
//  --go_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
//  --go-grpc_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
//  --grpc-gateway_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
//  --openapiv2_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
//  /home/vtoroy/GolandProjects/profile_service/proto/newproto.proto

func main() {
	// Загружаем конфигурацию (переменные окружения)
	config.LoadConfig()

	// Подключаемся к Auth-сервису
	authClient := createAuthClient("localhost:50051") // расскомментировать

	// Подключение к базе данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Выполнение миграций
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		// Укажите домен, с которого разрешены запросы (например, ваш фронтенд)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // Для авторизации через cookie
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// go get -u github.com/swaggo/gin-swagger
	// go get -u github.com/swaggo/files

	router.Static("/swagger", "/home/vtoroy/GolandProjects/profile_service/internal/swaggerui/dist/")

	// Подключаем middleware
	//r.Use(middleware.AuthMiddleware(authClient)) // раскомментировать

	profilesGroup := router.Group("/api/profiles", middleware.AuthMiddleware(authClient))
	{
		profilesGroup.POST("/create", rProfile.CreateProfile(db))
		profilesGroup.GET("/:id", rProfile.GetProfile(db))
		profilesGroup.GET("/token", rProfile.GetProfileByToken(db))
	}

	//authClient := initAuthClient() // Инициализация gRPC-клиента для AuthService
	profileService := ps.NewProfileService(db, authClient)
	v1 := router.Group("/v1")
	{
		v1.GET("/profiles/token", profileService.GetProfileByTokenHTTP)            // Получение профиля по токену
		v1.GET("/profiles/userid/:user_id", profileService.GetProfileByUserIDHTTP) // Получение профиля по user_id
		v1.GET("/profiles/id/:id", profileService.GetProfileByIDHTTP)              // Получение профиля по id
		v1.POST("/profiles", profileService.CreateProfileHTTP)                     // Создание профиля
		v1.GET("/profiles", profileService.GetProfilesHTTP)                        // Получение списка профилей
	}
	// gRPC маршруты
	go func() {

		grpcServer := grpc.NewServer()

		// Создание сервиса для работы с профилями
		annotation_proto.RegisterProfileServiceServer(grpcServer, profileService) // Регистрируем ProfileServer для gRPC

		// Включение reflection для gRPC
		reflection.Register(grpcServer)

		listener, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen on port 50052: %v", err)
		}

		log.Println("gRPC server is running on port 50052")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}

	}()

	go func() {
		if err := router.Run("127.0.0.1:8082"); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	select {} // Программа не завершится и будет продолжать работать
}

func createAuthClient(address string) authpb.AuthServiceClient {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к Auth-сервису: %v", err)
	}
	return authpb.NewAuthServiceClient(conn)
}

//func initAuthClient() authProto.AuthServiceClient {
//	conn, err := grpc.Dial("auth_service_address:port", grpc.WithInsecure()) // Укажите адрес сервиса авторизации
//	if err != nil {
//		log.Fatalf("Failed to connect to auth service: %v", err)
//	}
//	return authProto.NewAuthServiceClient(conn)
//}
