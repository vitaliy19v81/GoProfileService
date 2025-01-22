// /home/vtoroy/GolandProjects/profile_service/profile_service/profile.go
package profile

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"

	profileProto "profile_service/internal/services/annotation_proto"
	authProto "profile_service/internal/services/auth_proto"
	p "profile_service/services"
)

//protoc \
// -I /home/vtoroy/GolandProjects/profile_service \
// -I /home/vtoroy/GolandProjects/profile_service/proto \
// --go_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
// --go-grpc_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
// --grpc-gateway_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
// --openapiv2_out=/home/vtoroy/GolandProjects/profile_service/internal/services \
// /home/vtoroy/GolandProjects/profile_service/proto/newproto.proto

type ProfileServiceServer struct {
	profileProto.UnimplementedProfileServiceServer                             // Встраивание gRPC-сервера с пустой реализацией
	AuthClient                                     authProto.AuthServiceClient // gRPC клиент авторизации
	DB                                             *sql.DB
}

type CreateProfileRequest struct {
	UserId         string                 `json:"userId" binding:"required"` // Тег должен совпадать с JSON
	FirstName      string                 `json:"firstName" binding:"required"`
	LastName       string                 `json:"lastName" binding:"required"`
	MiddleName     string                 `json:"middleName,omitempty"`
	Phone          string                 `json:"phone,omitempty"`
	Address        string                 `json:"address,omitempty"`
	Birthday       string                 `json:"birthday,omitempty"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

func NewProfileService(db *sql.DB, authClient authProto.AuthServiceClient) *ProfileServiceServer {
	if db == nil {
		log.Fatal("Database connection is nil")
	}
	if authClient == nil {
		log.Fatal("AuthServiceClient is nil")
	}
	return &ProfileServiceServer{
		DB:         db,
		AuthClient: authClient,
	}
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
	query := `SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at 
	          FROM profiles WHERE user_id = $1`

	profile, err = fetchProfileFromDB(ctx, s.DB, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found for id: %s", userID)
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &profileProto.GetProfileByTokenResponse{
		Profile: convertProfileToProto(profile),
	}, nil
}

func (s *ProfileServiceServer) CreateProfile(ctx context.Context, req *profileProto.CreateProfileRequest) (*profileProto.CreateProfileResponse, error) {
	profile := p.Profile{
		UserID:     req.UserId,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: stringToNullString(req.MiddleName),
		Phone:      stringToNullString(req.Phone),
		Address:    stringToNullString(req.Address),
		Birthday:   stringToNullTime(req.Birthday),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	additionalData, err := json.Marshal(req.AdditionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize additional data: %v", err)
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}

	query := `
	INSERT INTO profiles (user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`
	err = s.DB.QueryRowContext(ctx, query,
		profile.UserID, profile.FirstName, profile.LastName, profile.MiddleName,
		profile.Phone, profile.Address, profile.Birthday, additionalData,
		profile.CreatedAt, profile.UpdatedAt,
	).Scan(&profile.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %v", err)
	}

	return &profileProto.CreateProfileResponse{Id: profile.ID}, nil
}

func (s *ProfileServiceServer) GetProfile(ctx context.Context, req *profileProto.GetProfileRequest) (*profileProto.GetProfileResponse, error) {
	var profile p.Profile

	query := `
	SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
	FROM profiles WHERE id = $1`

	profile, err := fetchProfileFromDB(ctx, s.DB, query, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found for id: %s", req.Id)
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &profileProto.GetProfileResponse{
		Profile: convertProfileToProto(profile),
	}, nil
}

func (s *ProfileServiceServer) GetProfileByUserID(ctx context.Context, req *profileProto.GetProfileByUserIDRequest) (*profileProto.GetProfileByUserIDResponse, error) {
	var profile p.Profile

	query := `
	SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, additional_data, created_at, updated_at
	FROM profiles WHERE user_id = $1`

	profile, err := fetchProfileFromDB(ctx, s.DB, query, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profile not found for UserId: %s", req.UserId)
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &profileProto.GetProfileByUserIDResponse{
		Profile: convertProfileToProto(profile),
	}, nil
}

func (s *ProfileServiceServer) GetProfiles(ctx context.Context, req *profileProto.GetProfilesRequest) (*profileProto.GetProfilesResponse, error) {

	var profiles []p.Profile
	var totalRecords int

	// Считаем общее количество записей
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM profiles`).Scan(&totalRecords)
	if err != nil {
		return nil, err
	}

	// Запрос с лимитом и смещением
	rows, err := s.DB.Query(`
		SELECT id, user_id, first_name, last_name, middle_name, phone, address, birthday, created_at, updated_at
		FROM profiles LIMIT $1 OFFSET $2`, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка строк
	for rows.Next() {
		var profile p.Profile
		err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.FirstName,
			&profile.LastName,
			&profile.MiddleName,
			&profile.Phone,
			&profile.Address,
			&profile.Birthday,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		profiles = append(profiles, profile)
	}

	// Преобразуем в формат protobuf
	//grpcProfiles := make([]*profileProto.Profile, len(profiles))
	//for i, profile := range profiles {
	//	grpcProfiles[i] = &profileProto.Profile{
	//		Id:         profile.ID,
	//		UserId:     profile.UserID,
	//		FirstName:  profile.FirstName,
	//		LastName:   profile.LastName,
	//		MiddleName: nullStringToString(profile.MiddleName),
	//		Phone:      nullStringToString(profile.Phone),
	//		Address:    nullStringToString(profile.Address),
	//		Birthday:   nullTimeToString(profile.Birthday),
	//		CreatedAt:  profile.CreatedAt.Format(time.RFC3339),
	//		UpdatedAt:  profile.UpdatedAt.Format(time.RFC3339),
	//	}
	//}

	// Преобразуем в формат protobuf
	grpcProfiles := make([]*profileProto.Profile, len(profiles))
	for i, profile := range profiles {
		grpcProfiles[i] = convertProfileToProto(profile)
	}

	return &profileProto.GetProfilesResponse{
		Profiles:     grpcProfiles,
		TotalRecords: int32(totalRecords),
	}, nil
}

func (s *ProfileServiceServer) DeleteProfileByID(ctx context.Context, req *profileProto.DeleteProfileByIDRequest) (*profileProto.DeleteProfileByIDResponse, error) {

	_, err := s.DB.ExecContext(ctx, "DELETE FROM profiles WHERE id = $1", req.Id)
	if err != nil {
		return nil, err
	}

	return &profileProto.DeleteProfileByIDResponse{
		Message: "Profile deleted",
	}, nil
}

func (s *ProfileServiceServer) UpdateProfile(ctx context.Context, req *profileProto.UpdateProfileRequest) (*profileProto.UpdateProfileResponse, error) {
	query := `
		UPDATE profiles 
		SET first_name = $1, last_name = $2, middle_name = $3, phone = $4, address = $5, birthday = $6, additional_data = $7, updated_at = $8 
		WHERE user_id = $9
	`

	additionalData, err := json.Marshal(req.AdditionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize additional data: %v", err)
	}

	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, fmt.Errorf("invalid UUID: %v", err)
	}

	_, err = s.DB.ExecContext(ctx, query,
		req.FirstName, req.LastName, stringToNullString(req.MiddleName), stringToNullString(req.Phone),
		stringToNullString(req.Address), stringToNullTime(req.Birthday), additionalData, time.Now(), req.UserId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %v", err)
	}

	return &profileProto.UpdateProfileResponse{Message: "Profile updated successfully"}, nil
}

// nullStringToString конвертирует sql.NullString в обычную строку.
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// nullTimeToString конвертирует sql.NullTime в строку формата времени.
func nullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02T15:04:05-07:00")
	}
	return ""
}

// nullTimeToPtr преобразует sql.NullTime в *time.Time.
func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func stringToNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	// Попробуем распарсить строку в time.Time
	parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", s)
	if err != nil {
		return sql.NullTime{Valid: false} // Вернем невалидное значение, если не удалось распарсить
	}
	return sql.NullTime{Time: parsedTime, Valid: true}
}

func convertProfileToProto(profile p.Profile) *profileProto.Profile {
	var additionalData map[string]string
	if profile.AdditionalData != nil {
		// Пробуем распарсить RawMessage в map[string]string
		if err := json.Unmarshal(*profile.AdditionalData, &additionalData); err != nil {
			log.Println("Error unmarshalling AdditionalData:", err)
			additionalData = nil // Если не получилось распарсить, устанавливаем в nil
		}
	}

	return &profileProto.Profile{
		Id:             profile.ID,
		UserId:         profile.UserID,
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
		MiddleName:     nullStringToString(profile.MiddleName),
		Phone:          nullStringToString(profile.Phone),
		Address:        nullStringToString(profile.Address),
		Birthday:       nullTimeToString(profile.Birthday),
		AdditionalData: additionalData,                         // Если AdditionalData == nil, то здесь будет пустая строка
		CreatedAt:      profile.CreatedAt.Format(time.RFC3339), // форматирует time.Time в строку в формате ISO 8601,
		UpdatedAt:      profile.UpdatedAt.Format(time.RFC3339), // который совместим с JSON и протобуф
	}
}

func fetchProfileFromDB(ctx context.Context, db *sql.DB, query string, args ...interface{}) (p.Profile, error) {
	var profile p.Profile
	var additionalData sql.NullString // Для хранения JSONB или NULL //var additionalData []byte
	err := db.QueryRowContext(ctx, query, args...).Scan(
		&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName,
		&profile.MiddleName, &profile.Phone, &profile.Address, &profile.Birthday,
		&additionalData, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return profile, err
	}

	//if len(additionalData) > 0 {
	//	if err := json.Unmarshal(additionalData, &profile.AdditionalData); err != nil {
	//		return profile, fmt.Errorf("error unmarshalling additional_data: %w", err)
	//	}
	//}

	// Обработка additional_data
	if additionalData.Valid {
		raw := json.RawMessage(additionalData.String)
		profile.AdditionalData = &raw
	} else {
		profile.AdditionalData = nil
	}

	return profile, nil
}

//func convertProfileToProto(profile p.Profile) *profileProto.Profile {
//	return &profileProto.Profile{
//		Id:             profile.ID,
//		UserId:         profile.UserID,
//		FirstName:      profile.FirstName,
//		LastName:       profile.LastName,
//		MiddleName:     nullStringToString(profile.MiddleName),
//		Phone:          nullStringToString(profile.Phone),
//		Address:        nullStringToString(profile.Address),
//		Birthday:       nullTimeToString(profile.Birthday),
//		AdditionalData: convertAdditionalData(profile.AdditionalData),
//		CreatedAt:      profile.CreatedAt.Format(time.RFC3339),
//		UpdatedAt:      profile.UpdatedAt.Format(time.RFC3339),
//	}
//}

//func convertAdditionalData(data map[string]interface{}) map[string]string {
//	converted := make(map[string]string)
//	for key, value := range data {
//		if str, ok := value.(string); ok {
//			converted[key] = str
//		} else {
//			converted[key] = fmt.Sprintf("%v", value) // Преобразуем в строку, если не строка
//		}
//	}
//	return converted
//}

//// flattenUser преобразует структуру в плоскую.
//func flattenProfile(profile p.Profile) map[string]interface{} {
//	return map[string]interface{}{
//		"id":              profile.ID,
//		"user_id":         profile.UserID,
//		"first_name":      profile.FirstName,
//		"last_name":       profile.LastName,
//		"middle_name":     profile.MiddleName,
//		"phone":           profile.Phone,
//		"address":         profile.Address,
//		"birthday":        profile.Birthday,
//		"additional_data": profile.AdditionalData,
//		"created_at":      profile.CreatedAt.Format(time.RFC3339),
//		"updated_at":      profile.UpdatedAt.Format(time.RFC3339),
//
//		//"username":            nullStringToString(user.Username),
//		//"email":               nullStringToString(user.Email),
//		////"phone":               nullStringToString(user.Phone),
//		//"role":                nullStringToString(user.Role),
//		//"status":              nullStringToString(user.Status),
//		//"password_updated_at": nullTimeToString(user.PasswordUpdatedAt),
//		//"created_at":          nullTimeToString(user.CreatedAt),
//		//"last_login":          nullTimeToString(user.LastLogin),
//	}
//}
