package crypto

import (
	"errors"

	"github.com/golang-jwt/jwt"
	"github.com/mitchellh/mapstructure"
)

var ErrJwtManagerSigning = errors.New("could not generate a new jwt")
var ErrJwtManagerEmptyToken = errors.New("could not verify an empty jwt")
var ErrJwtManagerEmptySecret = errors.New("could not sign/verify witn an empty secret")
var ErrJwtManagerEmptyDecodingBody = errors.New("could not decode a jwt. Got empty interface")
var ErrJwtManagerInvalidSigningMethod = errors.New("unexpected jwt signing method")
var ErrJwtManagerCastOrInvalidToken = errors.New("could not cast claims or invalid jwt")

type JwtManager interface {
	Sign(secret string, payload interface {
		Valid() error
	}) (string, error)
	Verify(secret, jwtToken string, body interface{}) error
}

type onlyofficeJwtManager struct {
	key []byte
}

func NewOnlyofficeJwtManager() JwtManager {
	return onlyofficeJwtManager{}
}

func (j onlyofficeJwtManager) Sign(secret string, payload interface {
	Valid() error
}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	ss, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", ErrJwtManagerSigning
	}

	return ss, nil
}

func (j onlyofficeJwtManager) Verify(secret, jwtToken string, body interface{}) error {
	if secret == "" {
		return ErrJwtManagerEmptySecret
	}

	if jwtToken == "" {
		return ErrJwtManagerEmptyToken
	}

	if body == nil {
		return ErrJwtManagerEmptyDecodingBody
	}

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrJwtManagerInvalidSigningMethod
		}

		return []byte(secret), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return ErrJwtManagerCastOrInvalidToken
	} else {
		return mapstructure.Decode(claims, body)
	}
}
