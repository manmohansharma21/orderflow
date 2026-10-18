package http

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type authHandler struct {
	db *pgxpool.Pool
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func newAuthHandler(db *pgxpool.Pool) *authHandler {
	return &authHandler{db: db}
}

func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func (h *authHandler) register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "valid email and password of at least 8 characters are required",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not hash password",
		})
		return
	}

	var userID string

	err = h.db.QueryRow(
		c.Request.Context(),
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, 'CUSTOMER')
		 RETURNING id`,
		email,
		string(hash),
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "could not register user; email may already be registered",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    userID,
		"email": email,
		"role":  "CUSTOMER",
	})
}

func (h *authHandler) login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "valid email and password are required",
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	var userID string
	var passwordHash string
	var role string

	err := h.db.QueryRow(
		c.Request.Context(),
		`SELECT id, password_hash, role
		 FROM users
		 WHERE email = $1`,
		email,
	).Scan(&userID, &passwordHash, &role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not authenticate user",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	secret := jwtSecret()
	if len(secret) < 32 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "JWT_SECRET is not configured correctly",
		})
		return
	}

	now := time.Now()

	claims := authClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not create token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
		"token_type":   "Bearer",
		"expires_in":   86400,
		"user": gin.H{
			"id":    userID,
			"email": email,
			"role":  role,
		},
	})
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "bearer token required",
			})
			return
		}

		secret := jwtSecret()
		if len(secret) < 32 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "JWT_SECRET is not configured correctly",
			})
			return
		}

		claims := &authClaims{}

		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}
				return secret, nil
			},
		)

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		if claims.UserID == "" || claims.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actualRole := c.GetString("user_role")

		if actualRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
