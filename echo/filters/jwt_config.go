package filters

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

type JwtCustomClaims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

type JwtUserClaims struct {
	UserId string `json:"user_id"`
	IP     string `json:"ip"`
	TS     int64  `json:"ts"`
	Sign   string `json:"sign"`
	jwt.StandardClaims
}

var (
	HS256_KEY = []byte("secret")
)

var jwtOptions map[string]interface{}

func skipRoute(ctx echo.Context) bool {
	if jwtOptions == nil {
		return false
	}
	return false
}

func GetJWTFromCookie() middleware.JWTConfig {
	return middleware.JWTConfig{
		Skipper:     skipRoute,
		Claims:      &JwtCustomClaims{},
		ContextKey:  "JwtToken",
		SigningKey:  HS256_KEY,
		TokenLookup: "cookie:jwt",
		//ErrorHandler: func(err error) error {
		//	retMsg:=utils.NewRetMsg(nil)
		//	retMsg.PackResult(map[string]interface{}{
		//		"redirect":"/customer/v1/user/tuyoo/login",
		//	})
		//	return &echo.HTTPError{
		//		Code:     http.StatusUnauthorized,
		//		Message:  "/customer/v1/user/tuyoo/login",
		//		Internal: err,
		//	}
		//},
	}
}

func GetJWT(ctx echo.Context) (*JwtCustomClaims, error) {
	if ctx == nil {
		return nil, errors.New("echo.Context is nil")
	}
	jwttoken := ctx.Get("JwtToken")
	if jwttoken == nil {
		return nil, errors.New("JwtToken is not found %v")
	}
	jwt_token, ok := jwttoken.(*jwt.Token)
	if !ok {
		return nil, errors.Errorf("JwtToken is not *jwt.Token %v", jwt_token)
	}
	return jwt_token.Claims.(*JwtCustomClaims), nil
}

func GetSDKJWT(ctx echo.Context) (*JwtUserClaims, error) {
	if ctx == nil {
		return nil, errors.New("echo.Context is nil")
	}
	jwttoken := ctx.Get("JwtToken")
	jwt_token, ok := jwttoken.(*jwt.Token)
	if !ok {
		return nil, errors.Errorf("JwtToken is not *jwt.Token %v", jwt_token)
	}
	return jwt_token.Claims.(*JwtUserClaims), nil
}
