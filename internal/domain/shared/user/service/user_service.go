package service

import (
	"context"
	"errors"

	"gin_template/internal/domain/shared/user/dto"
	"gin_template/internal/domain/shared/user/entity"
	"gin_template/internal/domain/shared/user/model"

	idmng "github.com/wiidz/goutil/mngs/identityMng"
	repoMng "github.com/wiidz/goutil/mngs/repoMng"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid login credentials")
)

type Service struct {
	users *repoMng.Repo[entity.UserEntity]
	auth  *idmng.IdentityMng
}

func New(users *repoMng.Repo[entity.UserEntity], auth *idmng.IdentityMng) *Service {
	return &Service{users: users, auth: auth}
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (dto.TokenPair, error) {
	device := req.Device
	if device == "" {
		device = "client"
	}

	user, err := s.findUser(ctx, req.LoginID)
	if err != nil {
		return dto.TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return dto.TokenPair{}, ErrInvalidCredentials
	}

	pair, err := s.auth.Login(ctx, req.LoginID, device)
	if err != nil {
		return dto.TokenPair{}, err
	}
	return dto.TokenPair{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken}, nil
}

func (s *Service) Logout(ctx context.Context) error { return s.auth.Logout(ctx) }

func (s *Service) CurrentLoginID(ctx context.Context) string { return s.auth.CurrentLoginID(ctx) }

func (s *Service) findUser(ctx context.Context, loginID string) (*model.User, error) {
	ue, err := s.users.First(ctx, repoMng.WithEq("login_id", loginID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if ue == nil {
		return nil, ErrInvalidCredentials
	}

	return &model.User{
		ID:           ue.ID,
		LoginID:      ue.LoginID,
		Nickname:     ue.Nickname,
		PasswordHash: ue.PasswordHash,
	}, nil
}
