package usecase

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/cnpf/feeder-backend/graph/model"
	"github.com/cnpf/feeder-backend/graph/scalars"
	"github.com/cnpf/feeder-backend/internal/auth"
	"github.com/cnpf/feeder-backend/internal/domain/entity"
	apperrors "github.com/cnpf/feeder-backend/internal/errors"
	"github.com/cnpf/feeder-backend/internal/repository/interface"
	"github.com/cnpf/feeder-backend/internal/validation"
)

// UseCaseImpl implements UseCase interface
// Uses repository interfaces (dependency inversion)
type UseCaseImpl struct {
	userRepo         repository.UserRepository
	reportRepo       repository.ReportRepository
	competitionRepo  repository.CompetitionRepository
	registrationRepo repository.RegistrationRepository
}

// NewUseCase creates a new use case implementation
func NewUseCase(
	userRepo repository.UserRepository,
	reportRepo repository.ReportRepository,
	competitionRepo repository.CompetitionRepository,
	registrationRepo repository.RegistrationRepository,
) UseCase {
	return &UseCaseImpl{
		userRepo:         userRepo,
		reportRepo:       reportRepo,
		competitionRepo:  competitionRepo,
		registrationRepo: registrationRepo,
	}
}

// Register implements UseCase.Register
func (u *UseCaseImpl) Register(ctx context.Context, email, username, password, passwordConfirm string, avatar *PhotoUpload) (*model.AuthResult, error) {
	// Validate input
	if err := validation.ValidateRegisterInput(email, username, password, passwordConfirm); err != nil {
		return nil, apperrors.WrapError("Неверные входные данные", err)
	}

	// Check if user exists
	_, err := u.userRepo.FindByEmailOrUsername(ctx, email, username)
	if err == nil {
		return nil, fmt.Errorf("Email или имя пользователя уже используются")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось захешировать пароль", err)
	}

	// Check if this is the first user (make admin)
	usersCount, err := u.userRepo.CountUsers(ctx)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось подсчитать количество пользователей", err)
	}
	isAdmin := usersCount == 0

	// Handle avatar upload
	var avatarData map[string]interface{}
	hasAvatar := false
	if avatar != nil {
		// Validate file type
		if !strings.HasPrefix(avatar.ContentType, "image/") {
			return nil, fmt.Errorf("Аватар должен быть изображением")
		}

		// Validate file size (max 512KB)
		const maxAvatarSize = 512 * 1024
		if avatar.Size > maxAvatarSize {
			return nil, fmt.Errorf("Файл аватара слишком большой (максимум 512KB)")
		}

		// Read file data
		data, err := io.ReadAll(avatar.File)
		if err != nil {
			return nil, apperrors.WrapError("Не удалось прочитать файл аватара", err)
		}

		avatarData = map[string]interface{}{
			"contentType": avatar.ContentType,
			"data":        data,
		}
		hasAvatar = true
	}

	// Create domain entity
	user := &entity.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		IsAdmin:      isAdmin,
		HasAvatar:    hasAvatar,
		Avatar:       avatarData,
		CreatedAt:    time.Now(),
	}

	// Save user
	userID, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать пользователя", err)
	}

	// Generate token
	token, err := auth.SignToken(userID, email)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать токен", err)
	}

	return &model.AuthResult{
		Ok:    true,
		Token: &token,
	}, nil
}

// Login implements UseCase.Login
func (u *UseCaseImpl) Login(ctx context.Context, login, password string) (*model.AuthResult, error) {
	// Validate input
	if err := validation.ValidateLoginInput(login, password); err != nil {
		return nil, apperrors.WrapError("Неверные входные данные", err)
	}

	// Find user
	user, err := u.userRepo.FindByEmailOrUsername(ctx, login, login)
	if err != nil {
		return nil, fmt.Errorf("Неверный email или пароль")
	}

	// Verify password
	if !auth.VerifyPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("Неверный email или пароль")
	}

	// Generate token
	token, err := auth.SignToken(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать токен", err)
	}

	return &model.AuthResult{
		Ok:    true,
		Token: &token,
	}, nil
}

// Logout implements UseCase.Logout
func (u *UseCaseImpl) Logout(ctx context.Context) (bool, error) {
	// Logout is handled by clearing cookie in resolver layer
	return true, nil
}

// GetCurrentUser implements UseCase.GetCurrentUser
// userID should be extracted from context in resolver layer and passed here
func (u *UseCaseImpl) GetCurrentUser(ctx context.Context, userID string) (*model.User, error) {
	if userID == "" {
		return nil, nil
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil
	}

	return entityToGraphQLUser(user), nil
}

// UpdateProfile implements UseCase.UpdateProfile
func (u *UseCaseImpl) UpdateProfile(ctx context.Context, userID string, username *string, removeAvatar *bool, avatar io.Reader, avatarSize int64, avatarContentType string) (*model.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("Не авторизован")
	}

	// Get current user
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}

	update := false
	if username != nil && *username != user.Username {
		// Check if username is already taken
		existing, err := u.userRepo.FindByEmailOrUsername(ctx, "", *username)
		if err == nil && existing.ID != user.ID {
			return nil, fmt.Errorf("Имя пользователя уже используется")
		}
		user.Username = *username
		update = true
	}

	if removeAvatar != nil && *removeAvatar {
		user.HasAvatar = false
		user.Avatar = nil
		update = true
	}

	if avatar != nil {
		// Validate file type
		if !strings.HasPrefix(avatarContentType, "image/") {
			return nil, fmt.Errorf("Аватар должен быть изображением")
		}

		// Validate file size (max 512KB)
		const maxAvatarSize = 512 * 1024
		if avatarSize > maxAvatarSize {
			return nil, fmt.Errorf("Файл аватара слишком большой (максимум 512KB)")
		}

		// Read file data
		data, err := io.ReadAll(avatar)
		if err != nil {
			return nil, apperrors.WrapError("Не удалось прочитать файл аватара", err)
		}

		user.Avatar = map[string]interface{}{
			"contentType": avatarContentType,
			"data":        data,
		}
		user.HasAvatar = true
		update = true
	}

	if !update {
		return nil, fmt.Errorf("Нет полей для обновления")
	}

	// Update user
	err = u.userRepo.Update(ctx, userID, user)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось обновить пользователя", err)
	}

	// Get updated user
	updatedUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти обновленного пользователя", err)
	}

	return entityToGraphQLUser(updatedUser), nil
}

// UpdatePassword implements UseCase.UpdatePassword
func (u *UseCaseImpl) UpdatePassword(ctx context.Context, userID string, oldPassword, newPassword string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("Не авторизован")
	}

	// Get current user
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("Пользователь не найден")
	}

	// Verify old password
	if !auth.VerifyPassword(oldPassword, user.PasswordHash) {
		return false, fmt.Errorf("Неверный текущий пароль")
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return false, apperrors.WrapError("Не удалось захешировать новый пароль", err)
	}

	// Update password
	user.PasswordHash = newPasswordHash
	err = u.userRepo.Update(ctx, userID, user)
	if err != nil {
		return false, apperrors.WrapError("Не удалось обновить пароль", err)
	}

	return true, nil
}

// Helper function to convert entity.User to model.User
func entityToGraphQLUser(e *entity.User) *model.User {
	if e == nil {
		return nil
	}

	var avatarURL *string
	if e.HasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", e.ID)
		avatarURL = &url
	}

	username := e.Username
	if username == "" {
		username = e.Email
	}

	return &model.User{
		ID:        e.ID,
		Email:     e.Email,
		Username:  username,
		IsAdmin:   e.IsAdmin,
		HasAvatar: e.HasAvatar,
		AvatarURL: avatarURL,
	}
}

// Helper function to convert entity.Report to model.Report
func (u *UseCaseImpl) entityToGraphQLReport(ctx context.Context, report *entity.Report, currentUserID string) (*model.Report, error) {
	if report == nil {
		return nil, nil
	}

	// Get author
	author, err := u.userRepo.FindByID(ctx, report.AuthorID)
	if err != nil {
		// If author not found, create a default author
		author = &entity.User{
			ID:        report.AuthorID,
			Username:  "unknown",
			HasAvatar: false,
		}
	}

	// Format author
	authorUsername := author.Username
	if authorUsername == "" {
		authorUsername = author.Email
	}
	var authorAvatarURL *string
	if author.HasAvatar {
		url := fmt.Sprintf("/api/user/avatar/%s", author.ID)
		authorAvatarURL = &url
	}

	// Format photos
	photos := make([]*model.Photo, len(report.Photos))
	for i := range report.Photos {
		photos[i] = &model.Photo{
			URL: fmt.Sprintf("/api/reports/%s/photos/%d", report.ID, i),
		}
	}

	// Format dates
	var createdAt *scalars.Time
	if !report.CreatedAt.IsZero() {
		t := scalars.Time(report.CreatedAt)
		createdAt = &t
	}

	var updatedAt *scalars.Time
	if !report.UpdatedAt.IsZero() {
		t := scalars.Time(report.UpdatedAt)
		updatedAt = &t
	}

	// Determine canEdit (user must be author or admin)
	canEdit := false
	if currentUserID != "" {
		currentUser, err := u.userRepo.FindByID(ctx, currentUserID)
		if err == nil {
			canEdit = currentUser.IsAdmin || currentUser.ID == report.AuthorID
		}
	}

	return &model.Report{
		ID:        report.ID,
		Title:     report.Title,
		Text:      report.Text,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		AuthorID:  report.AuthorID,
		Author: &model.Author{
			ID:        author.ID,
			Username:  authorUsername,
			HasAvatar: author.HasAvatar,
			AvatarURL: authorAvatarURL,
		},
		Photos:  photos,
		CanEdit: canEdit,
	}, nil
}

// Helper function to convert entity.Competition to model.Competition
func (u *UseCaseImpl) entityToGraphQLCompetition(competition *entity.Competition) (*model.Competition, error) {
	if competition == nil {
		return nil, nil
	}

	// Format tours
	tours := make([]*model.Tour, len(competition.Tours))
	for i, tour := range competition.Tours {
		tours[i] = &model.Tour{
			Date: scalars.Time(tour.Date),
			Time: tour.Time,
		}
	}

	// Format dates
	var startDate scalars.Time
	if competition.StartDate != nil {
		startDate = scalars.Time(*competition.StartDate)
	}

	var endDate scalars.Time
	if competition.EndDate != nil {
		endDate = scalars.Time(*competition.EndDate)
	}

	var openingDate *scalars.Time
	if competition.OpeningDate != nil {
		t := scalars.Time(*competition.OpeningDate)
		openingDate = &t
	}

	var createdAt *scalars.Time
	if !competition.CreatedAt.IsZero() {
		t := scalars.Time(competition.CreatedAt)
		createdAt = &t
	}

	var updatedAt *scalars.Time
	if !competition.UpdatedAt.IsZero() {
		t := scalars.Time(competition.UpdatedAt)
		updatedAt = &t
	}

	return &model.Competition{
		ID:               competition.ID,
		Title:            competition.Title,
		StartDate:        startDate,
		EndDate:          endDate,
		Location:         competition.Location,
		Tours:            tours,
		OpeningDate:      openingDate,
		OpeningTime:      competition.OpeningTime,
		IndividualFormat: competition.IndividualFormat,
		TeamFormat:       competition.TeamFormat,
		Fee:              competition.Fee,
		TeamLimit:        competition.TeamLimit,
		Regulations:      competition.Regulations,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// GetReports implements UseCase.GetReports
func (u *UseCaseImpl) GetReports(ctx context.Context, currentUserID string, limit *int) ([]*model.Report, error) {
	reportLimit := 20
	if limit != nil {
		if *limit < 1 {
			reportLimit = 1
		} else if *limit > 30 {
			reportLimit = 30
		} else {
			reportLimit = *limit
		}
	}

	reports, err := u.reportRepo.FindAll(ctx, reportLimit)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось получить отчеты", err)
	}

	result := make([]*model.Report, 0, len(reports))
	for _, report := range reports {
		graphQLReport, err := u.entityToGraphQLReport(ctx, report, currentUserID)
		if err != nil {
			continue // Skip reports with errors
		}
		result = append(result, graphQLReport)
	}

	return result, nil
}

// GetReport implements UseCase.GetReport
func (u *UseCaseImpl) GetReport(ctx context.Context, currentUserID string, id string) (*model.Report, error) {
	report, err := u.reportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Отчет не найден")
	}

	return u.entityToGraphQLReport(ctx, report, currentUserID)
}

// CreateReport implements UseCase.CreateReport
func (u *UseCaseImpl) CreateReport(ctx context.Context, userID string, title, text string, photos []*PhotoUpload) (*model.Report, error) {
	if userID == "" {
		return nil, fmt.Errorf("Не авторизован")
	}

	// Validate input
	title = strings.TrimSpace(title)
	text = strings.TrimSpace(text)

	if len(title) < 3 || len(title) > 120 {
		return nil, fmt.Errorf("Заголовок должен быть от 3 до 120 символов")
	}
	if len(text) < 1 || len(text) > 5000 {
		return nil, fmt.Errorf("Текст должен быть от 1 до 5000 символов")
	}

	// Process photo uploads
	photosList := make([]interface{}, 0)
	if len(photos) > 0 {
		const maxPhotos = 10
		const maxPhotoSize = 2 * 1024 * 1024 // 2MB

		if len(photos) > maxPhotos {
			return nil, fmt.Errorf("Слишком много фотографий (макс %d)", maxPhotos)
		}

		for _, upload := range photos {
			if upload == nil {
				continue
			}

			// Validate file type
			if !strings.HasPrefix(upload.ContentType, "image/") {
				return nil, fmt.Errorf("Фотография должна быть изображением")
			}

			// Validate file size
			if upload.Size > maxPhotoSize {
				return nil, fmt.Errorf("Файл фотографии слишком большой (макс 2МБ)")
			}

			// Read file data
			data, err := io.ReadAll(upload.File)
			if err != nil {
				return nil, apperrors.WrapError("Не удалось прочитать файл фотографии", err)
			}

			photosList = append(photosList, map[string]interface{}{
				"contentType": upload.ContentType,
				"data":        data,
			})
		}
	}

	// Create domain entity
	reportEntity := &entity.Report{
		AuthorID:  userID,
		Title:     title,
		Text:      text,
		Photos:    photosList,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	reportID, err := u.reportRepo.Create(ctx, reportEntity)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать отчет", err)
	}

	// Get created report
	createdReport, err := u.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти созданный отчет", err)
	}

	return u.entityToGraphQLReport(ctx, createdReport, userID)
}

// UpdateReport implements UseCase.UpdateReport
func (u *UseCaseImpl) UpdateReport(ctx context.Context, userID string, id string, title, text *string, removePhoto []int, removeAllPhotos *bool, photos []*PhotoUpload) (*model.Report, error) {
	if userID == "" {
		return nil, fmt.Errorf("Не авторизован")
	}

	// Check if report exists and get author
	authorID, err := u.reportRepo.GetAuthorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Отчет не найден")
	}

	// Check permissions (user must be author or admin)
	currentUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}

	if !currentUser.IsAdmin && currentUser.ID != authorID {
		return nil, fmt.Errorf("Доступ запрещен")
	}

	// Get existing report
	reportDoc, err := u.reportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Отчет не найден")
	}

	// Start with existing report data
	updatedReport := &entity.Report{
		ID:        reportDoc.ID,
		AuthorID:  reportDoc.AuthorID,
		Title:     reportDoc.Title,
		Text:      reportDoc.Text,
		Photos:    reportDoc.Photos,
		CreatedAt: reportDoc.CreatedAt,
		UpdatedAt: time.Now(),
	}

	update := false

	// Update title if provided
	if title != nil {
		titleTrimmed := strings.TrimSpace(*title)
		if len(titleTrimmed) < 3 || len(titleTrimmed) > 120 {
			return nil, fmt.Errorf("Заголовок должен быть от 3 до 120 символов")
		}
		updatedReport.Title = titleTrimmed
		update = true
	}

	// Update text if provided
	if text != nil {
		textTrimmed := strings.TrimSpace(*text)
		if len(textTrimmed) < 1 || len(textTrimmed) > 5000 {
			return nil, fmt.Errorf("Текст должен быть от 1 до 5000 символов")
		}
		updatedReport.Text = textTrimmed
		update = true
	}

	// Handle photo removal
	if removeAllPhotos != nil && *removeAllPhotos {
		updatedReport.Photos = []interface{}{}
		update = true
	} else if len(removePhoto) > 0 {
		removeIdx := make(map[int]bool)
		for _, idx := range removePhoto {
			removeIdx[idx] = true
		}
		newPhotos := []interface{}{}
		for i, photo := range updatedReport.Photos {
			if !removeIdx[i] {
				newPhotos = append(newPhotos, photo)
			}
		}
		updatedReport.Photos = newPhotos
		update = true
	}

	// Handle new photo uploads
	if len(photos) > 0 {
		const maxPhotos = 10
		const maxPhotoSize = 2 * 1024 * 1024 // 2MB

		currentCount := len(updatedReport.Photos)
		if currentCount+len(photos) > maxPhotos {
			return nil, fmt.Errorf("too many photos (max %d)", maxPhotos)
		}

		// Process new uploads
		for _, upload := range photos {
			if upload == nil {
				continue
			}

			// Validate file type
			if !strings.HasPrefix(upload.ContentType, "image/") {
				return nil, fmt.Errorf("Фотография должна быть изображением")
			}

			// Validate file size
			if upload.Size > maxPhotoSize {
				return nil, fmt.Errorf("Файл фотографии слишком большой (макс 2МБ)")
			}

			// Read file data
			data, err := io.ReadAll(upload.File)
			if err != nil {
				return nil, apperrors.WrapError("Не удалось прочитать файл фотографии", err)
			}

			updatedReport.Photos = append(updatedReport.Photos, map[string]interface{}{
				"contentType": upload.ContentType,
				"data":        data,
			})
		}
		update = true
	}

	if !update {
		return nil, fmt.Errorf("Нет полей для обновления")
	}

	// Update report
	err = u.reportRepo.Update(ctx, id, updatedReport)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось обновить отчет", err)
	}

	// Get updated report
	updatedReportDoc, err := u.reportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти обновленный отчет", err)
	}

	return u.entityToGraphQLReport(ctx, updatedReportDoc, userID)
}

// DeleteReport implements UseCase.DeleteReport
func (u *UseCaseImpl) DeleteReport(ctx context.Context, userID string, id string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("Не авторизован")
	}

	// Check if report exists and get author
	authorID, err := u.reportRepo.GetAuthorID(ctx, id)
	if err != nil {
		return false, fmt.Errorf("Отчет не найден")
	}

	// Check permissions (user must be author or admin)
	currentUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("пользователь не найден")
	}

	if !currentUser.IsAdmin && currentUser.ID != authorID {
		return false, fmt.Errorf("Доступ запрещен")
	}

	err = u.reportRepo.Delete(ctx, id)
	if err != nil {
		return false, apperrors.WrapError("не удалось удалить отчет", err)
	}

	return true, nil
}

// GetCompetitions implements UseCase.GetCompetitions
func (u *UseCaseImpl) GetCompetitions(ctx context.Context) ([]*model.Competition, error) {
	competitions, err := u.competitionRepo.FindAll(ctx)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось получить соревнования", err)
	}

	result := make([]*model.Competition, 0, len(competitions))
	for _, competition := range competitions {
		graphQLCompetition, err := u.entityToGraphQLCompetition(competition)
		if err != nil {
			continue // Skip competitions with errors
		}
		result = append(result, graphQLCompetition)
	}

	return result, nil
}

// GetCompetition implements UseCase.GetCompetition
func (u *UseCaseImpl) GetCompetition(ctx context.Context, id string) (*model.Competition, error) {
	competition, err := u.competitionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("соревнование не найдено")
	}

	return u.entityToGraphQLCompetition(competition)
}

// CreateCompetition implements UseCase.CreateCompetition
func (u *UseCaseImpl) CreateCompetition(ctx context.Context, input *model.CompetitionInput) (*model.Competition, error) {
	if input == nil {
		return nil, fmt.Errorf("Входные данные не могут быть пустыми")
	}

	if !input.IndividualFormat && !input.TeamFormat {
		return nil, fmt.Errorf("Выберите хотя бы один формат соревнований")
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("Неверная дата начала: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("Неверная дата окончания: %w", err)
	}

	// Convert tours to entity.Tour
	entityTours := make([]entity.Tour, len(input.Tours))
	for i, tourInput := range input.Tours {
		tourDate, err := time.Parse(time.RFC3339, tourInput.Date)
		if err != nil {
			return nil, fmt.Errorf("Неверная дата тура: %w", err)
		}
		entityTours[i] = entity.Tour{
			Date: tourDate,
			Time: tourInput.Time,
		}
	}

	var openingDateEntity *time.Time
	if input.OpeningDate != nil && *input.OpeningDate != "" {
		od, err := time.Parse(time.RFC3339, *input.OpeningDate)
		if err != nil {
			return nil, fmt.Errorf("Неверная дата открытия: %w", err)
		}
		openingDateEntity = &od
	}

	var fee *float64
	if input.Fee != nil && *input.Fee != "" {
		f, err := strconv.ParseFloat(*input.Fee, 64)
		if err != nil {
			return nil, fmt.Errorf("Неверная плата: %w", err)
		}
		fee = &f
	}

	var teamLimitInt *int
	if input.TeamLimit != nil && *input.TeamLimit != "" {
		tl, err := strconv.ParseInt(*input.TeamLimit, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Неверное ограничение команд: %w", err)
		}
		t := int(tl)
		teamLimitInt = &t
	}

	var regulationsStr *string
	if input.Regulations != nil {
		s := strings.TrimSpace(*input.Regulations)
		regulationsStr = &s
	}

	// Create entity.Competition
	competitionEntity := &entity.Competition{
		Title:            strings.TrimSpace(input.Title),
		StartDate:        &startDate,
		EndDate:          &endDate,
		Location:         strings.TrimSpace(input.Location),
		Tours:            entityTours,
		OpeningDate:      openingDateEntity,
		OpeningTime:      input.OpeningTime,
		IndividualFormat: input.IndividualFormat,
		TeamFormat:       input.TeamFormat,
		Fee:              fee,
		TeamLimit:        teamLimitInt,
		Regulations:      regulationsStr,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	competitionID, err := u.competitionRepo.Create(ctx, competitionEntity)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать соревнование", err)
	}

	// Get created competition
	createdCompetition, err := u.competitionRepo.FindByID(ctx, competitionID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти созданное соревнование", err)
	}

	return u.entityToGraphQLCompetition(createdCompetition)
}

// UpdateCompetition implements UseCase.UpdateCompetition
func (u *UseCaseImpl) UpdateCompetition(ctx context.Context, id string, input *model.CompetitionInput) (*model.Competition, error) {
	if input == nil {
		return nil, fmt.Errorf("Входные данные не могут быть пустыми")
	}

	if !input.IndividualFormat && !input.TeamFormat {
		return nil, fmt.Errorf("Выберите хотя бы один формат соревнований")
	}

	// Check if competition exists
	existingCompetition, err := u.competitionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Соревнование не найдено")
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("Неверная дата начала: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("Неверная дата окончания: %w", err)
	}

	// Convert tours to entity.Tour
	entityTours := make([]entity.Tour, len(input.Tours))
	for i, tourInput := range input.Tours {
		tourDate, err := time.Parse(time.RFC3339, tourInput.Date)
		if err != nil {
			return nil, fmt.Errorf("Неверная дата тура: %w", err)
		}
		entityTours[i] = entity.Tour{
			Date: tourDate,
			Time: tourInput.Time,
		}
	}

	var openingDateEntity *time.Time
	if input.OpeningDate != nil && *input.OpeningDate != "" {
		od, err := time.Parse(time.RFC3339, *input.OpeningDate)
		if err != nil {
			return nil, fmt.Errorf("Неверная дата открытия: %w", err)
		}
		openingDateEntity = &od
	}

	var fee *float64
	if input.Fee != nil && *input.Fee != "" {
		f, err := strconv.ParseFloat(*input.Fee, 64)
		if err != nil {
			return nil, fmt.Errorf("Неверная плата: %w", err)
		}
		fee = &f
	}

	var teamLimitInt *int
	if input.TeamLimit != nil && *input.TeamLimit != "" {
		tl, err := strconv.ParseInt(*input.TeamLimit, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Неверное ограничение команд: %w", err)
		}
		t := int(tl)
		teamLimitInt = &t
	}

	var regulationsStr *string
	if input.Regulations != nil {
		s := strings.TrimSpace(*input.Regulations)
		regulationsStr = &s
	}

	// Update entity.Competition
	updatedCompetition := &entity.Competition{
		ID:               existingCompetition.ID,
		Title:            strings.TrimSpace(input.Title),
		StartDate:        &startDate,
		EndDate:          &endDate,
		Location:         strings.TrimSpace(input.Location),
		Tours:            entityTours,
		OpeningDate:      openingDateEntity,
		OpeningTime:      input.OpeningTime,
		IndividualFormat: input.IndividualFormat,
		TeamFormat:       input.TeamFormat,
		Fee:              fee,
		TeamLimit:        teamLimitInt,
		Regulations:      regulationsStr,
		CreatedAt:        existingCompetition.CreatedAt,
		UpdatedAt:        time.Now(),
	}

	err = u.competitionRepo.Update(ctx, id, updatedCompetition)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось обновить соревнование", err)
	}

	// Get updated competition
	updatedCompetitionDoc, err := u.competitionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти обновленное соревнование", err)
	}

	return u.entityToGraphQLCompetition(updatedCompetitionDoc)
}

// DeleteCompetition implements UseCase.DeleteCompetition
func (u *UseCaseImpl) DeleteCompetition(ctx context.Context, id string) (bool, error) {
	err := u.competitionRepo.Delete(ctx, id)
	if err != nil {
		return false, apperrors.WrapError("Не удалось удалить соревнование", err)
	}

	return true, nil
}

// GetAdminUsers implements UseCase.GetAdminUsers
func (u *UseCaseImpl) GetAdminUsers(ctx context.Context) ([]*model.User, error) {
	users, err := u.userRepo.FindAll(ctx)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось получить пользователей", err)
	}

	result := make([]*model.User, 0, len(users))
	for _, user := range users {
		result = append(result, entityToGraphQLUser(user))
	}

	return result, nil
}

// GetAdminUser implements UseCase.GetAdminUser
func (u *UseCaseImpl) GetAdminUser(ctx context.Context, id string) (*model.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}

	return entityToGraphQLUser(user), nil
}

// AdminUpdateUser implements UseCase.AdminUpdateUser
func (u *UseCaseImpl) AdminUpdateUser(ctx context.Context, id string, isAdmin *bool) (*model.User, error) {
	if isAdmin == nil {
		return nil, fmt.Errorf("Нет полей для обновления")
	}

	// Don't allow removing admin rights from the last admin
	if !*isAdmin {
		adminsCount, err := u.userRepo.CountAdmins(ctx)
		if err != nil {
			return nil, apperrors.WrapError("Не удалось подсчитать количество админов", err)
		}
		if adminsCount <= 1 {
			targetUser, err := u.userRepo.FindByID(ctx, id)
			if err == nil && targetUser.IsAdmin {
				return nil, fmt.Errorf("Нельзя убрать права у последнего админа")
			}
		}
	}

	// Get existing user
	existingUser, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}

	// Update only isAdmin field
	updatedUser := &entity.User{
		ID:           existingUser.ID,
		Email:        existingUser.Email,
		Username:     existingUser.Username,
		PasswordHash: existingUser.PasswordHash,
		IsAdmin:      *isAdmin,
		HasAvatar:    existingUser.HasAvatar,
		Avatar:       existingUser.Avatar,
		CreatedAt:    existingUser.CreatedAt,
	}

	err = u.userRepo.Update(ctx, id, updatedUser)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось обновить пользователя", err)
	}

	// Fetch updated user
	finalUser, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти обновленного пользователя", err)
	}

	return entityToGraphQLUser(finalUser), nil
}

// AdminDeleteUser implements UseCase.AdminDeleteUser
func (u *UseCaseImpl) AdminDeleteUser(ctx context.Context, id string) (bool, error) {
	err := u.userRepo.Delete(ctx, id)
	if err != nil {
		return false, apperrors.WrapError("Не удалось удалить пользователя", err)
	}

	return true, nil
}

// CreateRegistration implements UseCase.CreateRegistration
func (u *UseCaseImpl) CreateRegistration(ctx context.Context, userID string, competitionID string, registrationType string, teamName *string, participants []ParticipantInput, coach *CoachInput) (*model.Registration, error) {
	if userID == "" {
		return nil, fmt.Errorf("Не авторизован")
	}

	// Validate competition exists
	competition, err := u.competitionRepo.FindByID(ctx, competitionID)
	if err != nil {
		return nil, fmt.Errorf("Соревнование не найдено")
	}

	// Validate registration type
	regType := entity.RegistrationType(registrationType)
	if regType != entity.RegistrationTypeIndividual && regType != entity.RegistrationTypeTeam {
		return nil, fmt.Errorf("Неверный тип регистрации")
	}

	// Validate format matches competition
	if regType == entity.RegistrationTypeIndividual && !competition.IndividualFormat {
		return nil, fmt.Errorf("Индивидуальный формат не поддерживается этим соревнованием")
	}
	if regType == entity.RegistrationTypeTeam && !competition.TeamFormat {
		return nil, fmt.Errorf("Командный формат не поддерживается этим соревнованием")
	}

	// Validate participants
	if regType == entity.RegistrationTypeIndividual {
		if len(participants) != 1 {
			return nil, fmt.Errorf("Для индивидуальной регистрации нужен один участник")
		}
		if teamName != nil {
			return nil, fmt.Errorf("Название команды не требуется для индивидуальной регистрации")
		}
		if coach != nil {
			return nil, fmt.Errorf("Тренер не требуется для индивидуальной регистрации")
		}
	} else {
		if len(participants) != 3 {
			return nil, fmt.Errorf("Для командной регистрации нужно три участника")
		}
		if teamName == nil || strings.TrimSpace(*teamName) == "" {
			return nil, fmt.Errorf("Название команды обязательно")
		}
	}

	// Validate participant names
	for i, p := range participants {
		if strings.TrimSpace(p.FirstName) == "" {
			return nil, fmt.Errorf("Имя участника %d обязательно", i+1)
		}
		if strings.TrimSpace(p.LastName) == "" {
			return nil, fmt.Errorf("Фамилия участника %d обязательна", i+1)
		}
	}

	// Check if user already registered (admins can register multiple times)
	currentUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}
	
	// Only check for existing registration if user is not admin
	if !currentUser.IsAdmin {
		existing, err := u.registrationRepo.FindByCompetitionAndUser(ctx, competitionID, userID)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("Вы уже зарегистрированы на это соревнование")
		}
	}

	// Convert participants
	entityParticipants := make([]entity.Participant, len(participants))
	for i, p := range participants {
		entityParticipants[i] = entity.Participant{
			FirstName: strings.TrimSpace(p.FirstName),
			LastName:  strings.TrimSpace(p.LastName),
		}
	}

	var entityCoach *entity.Coach
	if coach != nil {
		if strings.TrimSpace(coach.FirstName) == "" || strings.TrimSpace(coach.LastName) == "" {
			return nil, fmt.Errorf("Имя и фамилия тренера обязательны")
		}
		entityCoach = &entity.Coach{
			FirstName: strings.TrimSpace(coach.FirstName),
			LastName:  strings.TrimSpace(coach.LastName),
		}
	}

	// Create registration entity
	registration := &entity.Registration{
		CompetitionID: competitionID,
		UserID:        userID,
		Type:          regType,
		TeamName:      teamName,
		Participants:  entityParticipants,
		Coach:         entityCoach,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	registrationID, err := u.registrationRepo.Create(ctx, registration)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось создать регистрацию", err)
	}

	// Get created registration
	createdReg, err := u.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти созданную регистрацию", err)
	}

	return u.entityToGraphQLRegistration(createdReg, userID), nil
}

// GetRegistrationsByCompetition implements UseCase.GetRegistrationsByCompetition
func (u *UseCaseImpl) GetRegistrationsByCompetition(ctx context.Context, competitionID string, currentUserID string) ([]*model.Registration, error) {
	registrations, err := u.registrationRepo.FindByCompetitionID(ctx, competitionID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось получить регистрации", err)
	}

	result := make([]*model.Registration, 0, len(registrations))
	for _, reg := range registrations {
		result = append(result, u.entityToGraphQLRegistration(reg, currentUserID))
	}

	return result, nil
}

// UpdateRegistration implements UseCase.UpdateRegistration
func (u *UseCaseImpl) UpdateRegistration(ctx context.Context, userID string, registrationID string, teamName *string, participants []ParticipantInput, coach *CoachInput) (*model.Registration, error) {
	if userID == "" {
		return nil, fmt.Errorf("Не авторизован")
	}

	// Get existing registration
	existingReg, err := u.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("Регистрация не найдена")
	}

	// Check permissions (user must be author or admin)
	currentUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не найден")
	}

	if !currentUser.IsAdmin && currentUser.ID != existingReg.UserID {
		return nil, fmt.Errorf("Доступ запрещен")
	}

	// Validate participants count based on type
	if existingReg.Type == entity.RegistrationTypeIndividual {
		if len(participants) != 1 {
			return nil, fmt.Errorf("Для индивидуальной регистрации нужен один участник")
		}
	} else {
		if len(participants) != 3 {
			return nil, fmt.Errorf("Для командной регистрации нужно три участника")
		}
		if teamName == nil || strings.TrimSpace(*teamName) == "" {
			return nil, fmt.Errorf("Название команды обязательно")
		}
	}

	// Validate participant names
	for i, p := range participants {
		if strings.TrimSpace(p.FirstName) == "" {
			return nil, fmt.Errorf("Имя участника %d обязательно", i+1)
		}
		if strings.TrimSpace(p.LastName) == "" {
			return nil, fmt.Errorf("Фамилия участника %d обязательна", i+1)
		}
	}

	// Convert participants
	entityParticipants := make([]entity.Participant, len(participants))
	for i, p := range participants {
		entityParticipants[i] = entity.Participant{
			FirstName: strings.TrimSpace(p.FirstName),
			LastName:  strings.TrimSpace(p.LastName),
		}
	}

	var entityCoach *entity.Coach
	if coach != nil {
		if strings.TrimSpace(coach.FirstName) == "" || strings.TrimSpace(coach.LastName) == "" {
			return nil, fmt.Errorf("Имя и фамилия тренера обязательны")
		}
		entityCoach = &entity.Coach{
			FirstName: strings.TrimSpace(coach.FirstName),
			LastName:  strings.TrimSpace(coach.LastName),
		}
	}

	// Update registration
	updatedReg := &entity.Registration{
		ID:            existingReg.ID,
		CompetitionID: existingReg.CompetitionID,
		UserID:        existingReg.UserID,
		Type:          existingReg.Type,
		TeamName:      teamName,
		Participants:  entityParticipants,
		Coach:         entityCoach,
		CreatedAt:     existingReg.CreatedAt,
		UpdatedAt:     time.Now(),
	}

	err = u.registrationRepo.Update(ctx, registrationID, updatedReg)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось обновить регистрацию", err)
	}

	// Get updated registration
	updatedRegDoc, err := u.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, apperrors.WrapError("Не удалось найти обновленную регистрацию", err)
	}

	return u.entityToGraphQLRegistration(updatedRegDoc, userID), nil
}

// DeleteRegistration implements UseCase.DeleteRegistration
func (u *UseCaseImpl) DeleteRegistration(ctx context.Context, userID string, registrationID string) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("Не авторизован")
	}

	// Get existing registration
	existingReg, err := u.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return false, fmt.Errorf("Регистрация не найдена")
	}

	// Check permissions (user must be author or admin)
	currentUser, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("Пользователь не найден")
	}

	if !currentUser.IsAdmin && currentUser.ID != existingReg.UserID {
		return false, fmt.Errorf("Доступ запрещен")
	}

	err = u.registrationRepo.Delete(ctx, registrationID)
	if err != nil {
		return false, apperrors.WrapError("Не удалось удалить регистрацию", err)
	}

	return true, nil
}

// Helper function to convert entity.Registration to model.Registration
func (u *UseCaseImpl) entityToGraphQLRegistration(e *entity.Registration, currentUserID string) *model.Registration {
	if e == nil {
		return nil
	}

	participants := make([]*model.Participant, len(e.Participants))
	for i, p := range e.Participants {
		participants[i] = &model.Participant{
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}

	var coach *model.Coach
	if e.Coach != nil {
		coach = &model.Coach{
			FirstName: e.Coach.FirstName,
			LastName:  e.Coach.LastName,
		}
	}

	// Determine canEdit (user must be author or admin)
	canEdit := false
	if currentUserID != "" {
		ctx := context.Background()
		currentUser, err := u.userRepo.FindByID(ctx, currentUserID)
		if err == nil {
			canEdit = currentUser.IsAdmin || currentUser.ID == e.UserID
		}
	}

	return &model.Registration{
		ID:            e.ID,
		CompetitionID: e.CompetitionID,
		UserID:        e.UserID,
		Type:          string(e.Type),
		TeamName:      e.TeamName,
		Participants:  participants,
		Coach:         coach,
		CanEdit:       canEdit,
		CreatedAt:     scalars.Time(e.CreatedAt),
		UpdatedAt:     scalars.Time(e.UpdatedAt),
	}
}
