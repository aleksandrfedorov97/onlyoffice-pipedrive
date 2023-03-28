package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/port"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
)

var _ErrOperationTimeout = errors.New("operation timeout")

type userService struct {
	adapter   port.UserAccessServiceAdapter
	encryptor crypto.Encryptor
	logger    plog.Logger
}

func NewUserService(
	adapter port.UserAccessServiceAdapter,
	encryptor crypto.Encryptor,
	logger plog.Logger,
) port.UserAccessService {
	return userService{
		adapter:   adapter,
		encryptor: encryptor,
		logger:    logger,
	}
}

func (s userService) CreateUser(ctx context.Context, user domain.UserAccess) error {
	s.logger.Debugf("validating user %s to perform a persist action", user.ID)
	if err := user.Validate(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	errChan := make(chan error, 2)
	atokenChan := make(chan string, 1)
	rtokenChan := make(chan string, 1)

	go func() {
		defer wg.Done()
		aToken, err := s.encryptor.Encrypt(user.AccessToken)
		if err != nil {
			errChan <- err
			return
		}
		atokenChan <- aToken
	}()

	go func() {
		defer wg.Done()
		rToken, err := s.encryptor.Encrypt(user.RefreshToken)
		if err != nil {
			errChan <- err
			return
		}
		rtokenChan <- rToken
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return _ErrOperationTimeout
	default:
	}

	s.logger.Debugf("user %s is valid. Persisting to database: %s", user.ID, user.AccessToken)
	if err := s.adapter.InsertUser(ctx, domain.UserAccess{
		ID:           user.ID,
		AccessToken:  <-atokenChan,
		RefreshToken: <-rtokenChan,
		TokenType:    user.TokenType,
		Scope:        user.Scope,
		ExpiresAt:    user.ExpiresAt,
		ApiDomain:    user.ApiDomain,
	}); err != nil {
		return err
	}

	return nil
}

func (s userService) GetUser(ctx context.Context, uid string) (domain.UserAccess, error) {
	s.logger.Debugf("trying to select user with id: %s", uid)
	id := strings.TrimSpace(uid)

	if id == "" {
		return domain.UserAccess{}, &InvalidServiceParameterError{
			Name:   "UID",
			Reason: "Should not be blank",
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	errChan := make(chan error, 2)
	atokenChan := make(chan string, 1)
	rtokenChan := make(chan string, 1)

	user, err := s.adapter.SelectUserByID(ctx, id)
	if err != nil {
		return user, err
	}

	s.logger.Debugf("found a user: %v", user)

	go func() {
		defer wg.Done()
		aToken, err := s.encryptor.Decrypt(user.AccessToken)
		if err != nil {
			errChan <- err
			return
		}
		atokenChan <- aToken
	}()

	go func() {
		defer wg.Done()
		rToken, err := s.encryptor.Decrypt(user.RefreshToken)
		if err != nil {
			errChan <- err
			return
		}
		rtokenChan <- rToken
	}()

	wg.Wait()

	select {
	case err := <-errChan:
		return domain.UserAccess{}, err
	case <-ctx.Done():
		return domain.UserAccess{}, _ErrOperationTimeout
	default:
		return domain.UserAccess{
			ID:           user.ID,
			AccessToken:  <-atokenChan,
			RefreshToken: <-rtokenChan,
			TokenType:    user.TokenType,
			Scope:        user.Scope,
			ExpiresAt:    user.ExpiresAt,
			ApiDomain:    user.ApiDomain,
		}, nil
	}
}

func (s userService) UpdateUser(ctx context.Context, user domain.UserAccess) (domain.UserAccess, error) {
	s.logger.Debugf("validating user %s to perform an update action", user.ID)
	if err := user.Validate(); err != nil {
		return domain.UserAccess{}, err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	errChan := make(chan error, 2)
	atokenChan := make(chan string, 1)
	rtokenChan := make(chan string, 1)

	go func() {
		defer wg.Done()
		aToken, err := s.encryptor.Encrypt(user.AccessToken)
		if err != nil {
			errChan <- err
			return
		}
		atokenChan <- aToken
	}()

	go func() {
		defer wg.Done()
		rToken, err := s.encryptor.Encrypt(user.RefreshToken)
		if err != nil {
			errChan <- err
			return
		}
		rtokenChan <- rToken
	}()

	select {
	case err := <-errChan:
		return user, err
	case <-ctx.Done():
		return user, _ErrOperationTimeout
	default:
	}

	s.logger.Debugf("user %s is valid to perform an update action", user.ID)
	if _, err := s.adapter.UpsertUser(ctx, domain.UserAccess{
		ID:           user.ID,
		AccessToken:  <-atokenChan,
		RefreshToken: <-rtokenChan,
		TokenType:    user.TokenType,
		Scope:        user.Scope,
		ExpiresAt:    user.ExpiresAt,
		ApiDomain:    user.ApiDomain,
	}); err != nil {
		return user, err
	}

	return user, nil
}

func (s userService) DeleteUser(ctx context.Context, uid string) error {
	id := strings.TrimSpace(uid)
	s.logger.Debugf("validating uid %s to perform a delete action", id)

	if id == "" {
		return &InvalidServiceParameterError{
			Name:   "UID",
			Reason: "Should not be blank",
		}
	}

	s.logger.Debugf("uid %s is valid to perform a delete action", id)
	return s.adapter.DeleteUserByID(ctx, uid)
}
