package services

import (
	"errors"
	"medical-qa-assistant/internal/models"
	"medical-qa-assistant/internal/repositories"
	"medical-qa-assistant/internal/utils"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if username already exists
	_, err := s.userRepo.FindByUsername(req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	_, err = s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.RoleUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}
