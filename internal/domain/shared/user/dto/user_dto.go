package dto

import "github.com/wiidz/goutil/structs/networkStruct"

type LoginRequest struct {
	networkStruct.Params `swaggerignore:"true"`

	LoginID  string `json:"login_id" belong:"value" validate:"required"`
	Password string `json:"password" belong:"value" validate:"required"`
	Device   string `json:"device" belong:"value" default:"client"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
