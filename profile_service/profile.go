// /home/vtoroy/GolandProjects/profile_service/profile_service/profile.go
package profile

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	authProto "profile_service/internal/services/auth_proto"
	profileProto "profile_service/internal/services/proto"
	p "profile_service/services"
)

type ProfileServiceServer struct {
	profileProto.UnimplementedProfileServiceServer                             // Встраивание gRPC-сервера с пустой реализацией
	AuthClient                                     authProto.AuthServiceClient // gRPC клиент авторизации
	DB                                             *sql.DB
}

func (s *ProfileServiceServer) GetProfileByToken(ctx context.Context, req *profileProto.GetProfileByTokenRequest) (*profileProto.GetProfileByTokenResponse, error) {
	// Проверяем валидность токена через AuthClient
	validateReq := &authProto.ValidateTokenRequest{Token: req.Token}
	validateResp, err := s.AuthClient.ValidateToken(ctx, validateReq)
	if err != nil || !validateResp.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// Получаем user_id из ответа сервиса авторизации
	userID := validateResp.UserId

	// Запрашиваем профиль из базы данных
	var profile p.Profile
	query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, created_at, updated_at 
	          FROM profiles WHERE user_id = $1`
	err = s.DB.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName,
		&profile.MiddleName, &profile.Phone, &profile.Address, &profile.Birthday,
		&profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found for user_id: %s", userID)
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Формируем и возвращаем ответ
	return &profileProto.GetProfileByTokenResponse{
		Profile: &profileProto.Profile{
			Id:         profile.ID,
			UserId:     profile.UserID,
			FirstName:  profile.FirstName,
			LastName:   profile.LastName,
			MiddleName: profile.MiddleName,
			Phone:      profile.Phone,
			Address:    profile.Address,
			Birthday:   profile.Birthday,
			CreatedAt:  profile.CreatedAt.Format(time.RFC3339), // форматирует time.Time в строку в формате ISO 8601,
			UpdatedAt:  profile.UpdatedAt.Format(time.RFC3339), // который совместим с JSON и протобуф
		},
	}, nil
}

func NewProfileService(db *sql.DB) *ProfileServiceServer {
	if db == nil {
		log.Fatal("Database connection is nil")
	}
	return &ProfileServiceServer{DB: db}
}
