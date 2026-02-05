package errors

import "errors"

var AccessTokenExpiredError = errors.New("access token expired")
var RefreshTokenExpiredError = errors.New("refresh token expired")
