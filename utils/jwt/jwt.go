package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"soj_core/config"
)

type Claims struct {
	jwt.RegisteredClaims
	UID       uint64 `json:"uid"`
	SessionID uint64 `json:"session"`
	Type      string `json:"type"`
	JTI       string `json:"jti"`
}

// GenerateTokenPair generates a refresh token and its associated access token
// Returns: refreshToken, accessToken, error
func GenerateTokenPair(uid uint64, sessionID uint64, jti string) (string, string, error) {
	// Generate refresh token (1 month)
	refreshClaims := Claims{
		UID:       uid,
		SessionID: sessionID,
		Type:      "refresh",
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWTRefreshTokenDuration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(config.JWTPrivateKey())
	if err != nil {
		return "", "", err
	}

	// Generate access token (15 minutes) bound to refresh token
	accessClaims := Claims{
		UID:       uid,
		SessionID: sessionID,
		Type:      "access",
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWTAccessTokenDuration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, accessClaims)
	accessTokenString, err := accessToken.SignedString(config.JWTPrivateKey())
	if err != nil {
		return "", "", err
	}

	return refreshTokenString, accessTokenString, nil
}

func ParseJwtToken(tokenString string) (*Claims, error) {
	jwtClaims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims, func(token *jwt.Token) (interface{}, error) {
		return config.JWTPublicKey(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return &jwtClaims, nil
}

// HashToken creates a SHA-256 hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateJTI generates a unique token identifier
func GenerateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
