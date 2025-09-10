package controller 

import (
    "context"
	"errors"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"

    "proyecto/auth-server/pkg/model"
    
    
)

//error personal 
var ErrNotFound = errors.New("Not found")

type AuthRepository interface {
    GetHashByEmail(ctx context.Context, email string) (*model.AuthUser, error)
    Put(ctx context.Context, AuthUser *model.AuthUser) (error)
}

type Controller struct {
	repo AuthRepository
}

func New(repo AuthRepository) *Controller {
	return &Controller{repo: repo}
}


func (c *Controller) CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func (c *Controller) GenerateAccessToken(secret []byte, sub, email, role, jti string, ttl time.Duration) (string, error) {
    now := time.Now()
    claims := AccessClaims{
        Sub:   sub,
        Email: email,
        Role:  role,
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        jti,                      // jti
            Issuer:    "auth-service",          // iss (opcional)
            IssuedAt:  jwt.NewNumericDate(now), // iat
            ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), // exp
            // Audience: []string{"user-service"}, // aud (opcional)
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

// AccessClaims define los claims personalizados de un JWT de acceso.
type AccessClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func (c *Controller) GetHashByEmail(ctx context.Context, email string) (*model.AuthUser, error) {
	res, err := c.repo.GetHashByEmail(ctx, email)

	if err != nil {
		return nil, ErrNotFound
	}

	return res, err
}

func (c *Controller) Put(ctx context.Context, AuthUser *model.AuthUser) (*model.AuthUser, error) {
	err := c.repo.Put(ctx, AuthUser)
	if err != nil {
		return nil, err
	}
	return AuthUser, nil
}

type MetadataUser struct {
	Email              string    `json:"email"`
	FullName           string    `json:"full_name"`
	AvatarURL          string    `json:"avatar_url"`
	PhoneNumber        string    `json:"phone_number"`
	BirthDate          string 	 `json:"birth_date"`
	LastUpdated        string 	 `json:"last_updated"`
}