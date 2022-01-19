package middleware

import (
	"auth/domain"
	"auth/myerrors"
	"auth/user/delivery/response"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
)

// SecretMiddleware gets token secrets and ttl from .env and populates them to RequestCtx
func SecretMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		accessSecret := os.Getenv("ACCESS_SECRET")
		refreshSecret := os.Getenv("REFRESH_SECRET")

		if accessSecret == "" || refreshSecret == "" {
			log.Println("ERROR|Error retrieving secret")
			response.RespondInternalServerError(ctx)
			return
		}
		accessTtl := time.Second * 20
		refreshTtl := time.Minute * 10
		ctx.SetUserValue("accessSecret", accessSecret)
		ctx.SetUserValue("refreshSecret", refreshSecret)
		ctx.SetUserValue("accessTtl", accessTtl)
		ctx.SetUserValue("refreshTtl", refreshTtl)
		next(ctx)
	}
}

func CheckAuthMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	log.Println("INFO|CheckAuthMiddleware hit")
	return func(ctx *fasthttp.RequestCtx) {
		log.Println("INFO|Middleware hit")
		accessSecret, refreshSecret, err := GetSecretFromCtx(ctx)
		if err != nil {
			log.Println("ERROR|error retrieving secrets")
			response.RespondInternalServerError(ctx)
			return
		}

		access, err := ExtractToken(ctx, true)
		if err != nil {
			log.Println("ERROR|Extracting token:", err)
			ctx.SetStatusCode(fasthttp.StatusSeeOther)
			ctx.Response.Header.Add("Location", "/login")
			return
		}
		refresh, _ := ExtractToken(ctx, false)
		log.Println("INFO|Extracted following tokens:", access, "\n", refresh)
		IIN, isAdmin, err := ParseToken(ctx, access, true)
		if err != nil {
			if err.Error() == "Token is expired" {
				log.Println("ERROR|Access token expired, redirecting to update")
				ctx.Redirect("/update", fasthttp.StatusSeeOther)
				return
			}
			log.Printf("ERROR|Parse access token error: %v", err)
			ctx.SetStatusCode(fasthttp.StatusSeeOther)
			ctx.Response.Header.Add("Location", "/login")
			return
		}

		// Authorization success
		ctx.SetUserValue("accessSecret", accessSecret)
		ctx.SetUserValue("refreshSecret", refreshSecret)
		ctx.SetUserValue("access", access)
		ctx.SetUserValue("iin", IIN)
		ctx.SetUserValue("admin", isAdmin)
		SetValueFromToken(ctx, access)
		log.Println("INFO|Middleware auth passed")
		next(ctx)
	}
}

// SetValueFromToken sets user as ctx.UserValue
func SetValueFromToken(ctx *fasthttp.RequestCtx, token string) {
	claims := jwt.MapClaims{}
	p := jwt.Parser{}
	if _, _, err := p.ParseUnverified(token, &claims); err != nil {
		fmt.Println(err)
		return
	}
	user := domain.User{
		IIN:      claims["iin"].(string),
		Username: claims["username"].(string),
		Ts:       claims["createdAt"].(string),
		IsAdmin:  claims["admin"].(bool),
	}
	ctx.SetUserValue("user", user)
}

// extractCredential extracts login credentials
func extractCredential(ctx *fasthttp.RequestCtx) (login string, pass string) {
	return string(ctx.FormValue("login")), string(ctx.FormValue("password"))
}

// GetSecretFromCtx retrieves token secrets
func GetSecretFromCtx(ctx *fasthttp.RequestCtx) (accessSecret, refreshSecret string, err error) {
	accessSecret, accessOk := ctx.Value("accessSecret").(string)
	refreshSecret, refreshOk := ctx.Value("refreshSecret").(string)
	if accessOk && refreshOk {
		return
	}
	err = fmt.Errorf("Error getting secret")
	return
}

// ParseToken parses tokens and claims
func ParseToken(ctx *fasthttp.RequestCtx, token string, isAccess bool) (string, bool, error) {
	log.Println("INFO|ParseToken hit")
	accessSecret, refreshSecret, err := GetSecretFromCtx(ctx)
	if err != nil {
		return "", false, err
	}

	JWTToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Failed to extract token metadata, unexpected signing method: %v", token.Header["alg"])
		}
		if isAccess {
			return []byte(accessSecret), nil
		}
		return []byte(refreshSecret), nil
	})

	if err != nil {
		log.Printf("ERROR|Native parse err:%q", err.Error())
		return "", false, err
	}
	log.Println("INFO|Middleware: Extracting claims")
	claims, ok := JWTToken.Claims.(jwt.MapClaims)

	var IIN string
	var admin bool

	if ok && JWTToken.Valid {
		log.Println("INFO|Middlewrare: token ok & valid")
		admin, ok = claims["admin"].(bool)
		if !ok {
			return "", false, fmt.Errorf("Field admin not found")
		}
		_, ok = claims["username"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field username not found")
		}
		_, ok = claims["createdAt"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field userts not found")
		}

		IIN, ok = claims["iin"].(string)
		if !ok {
			return "", false, fmt.Errorf("Field iin not found")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return "", false, fmt.Errorf("Field exp not found")
		}

		expiredTime := time.Unix(int64(exp), 0)
		if time.Now().After(expiredTime) {
			return "", false, myerrors.ErrTokenExpired
		}
		log.Println("INFO|Middleware: everything ok, passing IIN and role from parsed token")
		return string(IIN), admin, nil
	}
	log.Println("INFO|Middleware: token not ok or not valid")
	return "", false, myerrors.ErrInvalidToken
}

// ExtractToken extracts access token value from cookies if isAccess is true and refresh token value otherwise
func ExtractToken(ctx *fasthttp.RequestCtx, isAccess bool) (string, error) {
	var token string
	token = string(ctx.Request.Header.Cookie("access"))
	if !isAccess {
		token = string(ctx.Request.Header.Cookie("refresh"))
	}
	if token == "" {
		return "", myerrors.ErrCookieNotFound
	}
	return token, nil
}
