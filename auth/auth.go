package auth

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"tarpaulin/datastore" // Add this import

	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
)

// Add these new structs for JWKS
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// AuthService handles authentication and authorization
type AuthService struct {
	Domain   string
	ClientID string
	Secret   string
}

// UserClaims represents the JWT claims we expect
type UserClaims struct {
	Sub string `json:"sub"`
	// Remove Email field since we only need sub
	jwt.StandardClaims
}

// NewAuthService creates a new Auth0 service
func NewAuthService() (*AuthService, error) {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	secret := os.Getenv("AUTH0_CLIENT_SECRET")

	if domain == "" || clientID == "" || secret == "" {
		return nil, errors.New("missing required Auth0 environment variables")
	}

	return &AuthService{
		Domain:   domain,
		ClientID: clientID,
		Secret:   secret,
	}, nil
}

// GenerateToken generates a JWT token
// Remove this entire function (lines 63-75)
// func (a *AuthService) GenerateToken(email string) (string, error) {
//     // Create the token
//     token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//         "sub":   email,
//         "email": email,
//         // Note: In a real Auth0 implementation, we would include more claims
//     })
//
//     // Sign the token with our secret
//     tokenString, err := token.SignedString([]byte(a.Secret))
//     if err != nil {
//         return "", err
//     }
//
//     return tokenString, nil
// }

// ValidateToken validates a JWT token with proper RS256 signature verification
func (a *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	fmt.Println("DEBUG: Starting proper JWT validation")

	// Parse the token with proper validation
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		fmt.Printf("DEBUG: Token kid: %s\n", kid)

		// Get the public key from Auth0's JWKS endpoint
		publicKey, err := a.getAuth0PublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %v", err)
		}

		return publicKey, nil
	})

	if err != nil {
		fmt.Printf("DEBUG: Token validation failed: %v\n", err)
		return nil, err
	}

	// Validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("DEBUG: Token signature is valid, validating claims")

		// Verify issuer
		expectedIssuer := fmt.Sprintf("https://%s/", a.Domain)
		if iss, ok := claims["iss"].(string); !ok || iss != expectedIssuer {
			return nil, fmt.Errorf("invalid issuer: expected %s, got %v", expectedIssuer, claims["iss"])
		}
		fmt.Printf("DEBUG: Issuer validated: %s\n", claims["iss"])

		// Verify audience
		if aud, ok := claims["aud"].(string); !ok || aud != a.ClientID {
			return nil, fmt.Errorf("invalid audience: expected %s, got %v", a.ClientID, claims["aud"])
		}
		fmt.Printf("DEBUG: Audience validated: %s\n", claims["aud"])

		// Verify expiration
		if exp, ok := claims["exp"].(float64); !ok {
			return nil, errors.New("exp claim not found or invalid")
		} else if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token expired")
		}
		fmt.Printf("DEBUG: Token expiration validated, expires at: %v\n", time.Unix(int64(claims["exp"].(float64)), 0))

		// Verify issued at time
		if iat, ok := claims["iat"].(float64); !ok {
			return nil, errors.New("iat claim not found or invalid")
		} else if time.Now().Unix() < int64(iat) {
			return nil, errors.New("token used before issued")
		}

		// Log all claims for debugging
		fmt.Println("DEBUG: All token claims:")
		for key, value := range claims {
			fmt.Printf("  %s: %v\n", key, value)
		}

		// Specifically log email and sub
		if email, ok := claims["email"].(string); ok {
			fmt.Printf("DEBUG: Found email claim: %s\n", email)
		}

		if sub, ok := claims["sub"].(string); ok {
			fmt.Printf("DEBUG: Found sub claim: %s\n", sub)
		}

		fmt.Println("DEBUG: Token validation successful")
		return token, nil
	}

	return nil, errors.New("invalid token claims")
}

// getAuth0PublicKey fetches the public key from Auth0's JWKS endpoint
func (a *AuthService) getAuth0PublicKey(kid string) (*rsa.PublicKey, error) {
	fmt.Printf("DEBUG: Fetching public key for kid: %s\n", kid)

	// Construct JWKS URL
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", a.Domain)
	fmt.Printf("DEBUG: JWKS URL: %s\n", jwksURL)

	// Fetch JWKS
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	// Parse JWKS
	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}

	fmt.Printf("DEBUG: Found %d keys in JWKS\n", len(jwks.Keys))

	// Find the key with matching kid
	for _, key := range jwks.Keys {
		fmt.Printf("DEBUG: Checking key with kid: %s\n", key.Kid)
		if key.Kid == kid {
			fmt.Printf("DEBUG: Found matching key for kid: %s\n", kid)

			// Convert the key to RSA public key
			if len(key.X5c) == 0 {
				return nil, errors.New("x5c certificate chain is empty")
			}

			// Parse the certificate
			certPEM := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", key.X5c[0])
			publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(certPEM))
			if err != nil {
				return nil, fmt.Errorf("failed to parse RSA public key: %v", err)
			}

			fmt.Println("DEBUG: Successfully parsed RSA public key")
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
}

// GetUserFromToken extracts user information from a token
// Update GetUserFromToken to only return sub
func (a *AuthService) GetUserFromToken(tokenString string) (string, error) {
	token, err := a.ValidateToken(tokenString)
	if err != nil {
		fmt.Printf("Token parsing failed: %v\n", err)
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Only get sub claim
		if sub, ok := claims["sub"].(string); ok {
			fmt.Printf("Using sub claim: %s\n", sub)
			return sub, nil
		}

		fmt.Println("No valid sub claim found in token")
		return "", errors.New("invalid token claims")
	}

	fmt.Println("Token claims could not be extracted")
	return "", errors.New("invalid token")
}

// AuthMiddleware is a Gin middleware that validates JWT tokens
func AuthMiddleware(authService *AuthService, userStore *datastore.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		fmt.Println("DEBUG: Authorization header:", authHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Authorization header is required"})
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Authorization header must be Bearer token"})
			return
		}

		fmt.Println("DEBUG: Extracting claims from token:", parts[1][:10]+"...")
		token, err := authService.ValidateToken(parts[1])
		if err != nil {
			fmt.Printf("DEBUG: Token parsing failed: %v\n", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
			return
		}

		// Extract claims from token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Get sub
			var sub string
			if subClaim, ok := claims["sub"].(string); ok {
				sub = subClaim
				fmt.Printf("DEBUG: Using sub claim: %s\n", sub)
			}

			if sub == "" {
				fmt.Println("DEBUG: No valid sub claim found in token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Invalid token claims"})
				return
			}

			// Use the userStore parameter directly instead of getting from context
			user, err := userStore.GetUserBySub(c.Request.Context(), sub)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "User not found"})
				return
			}

			// Set user object in context
			c.Set("user_sub", sub)
			c.Set("user", user)
			fmt.Println("DEBUG: AuthMiddleware successful, proceeding to next middleware")
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Invalid token"})
	}
}

// TODO: Implement AdminOnlyMiddleware and CourseInstructorMiddleware

// AuthenticateUser authenticates a user with Auth0 and returns a JWT token
func (a *AuthService) AuthenticateUser(username, password string) (string, error) {
	// Create the Auth0 authentication URL
	authURL := fmt.Sprintf("https://%s/oauth/token", a.Domain)

	// Prepare the request payload
	payload := map[string]string{
		"grant_type":    "password",
		"username":      username, // Use username instead of email
		"password":      password,
		"client_id":     a.ClientID,
		"client_secret": a.Secret,
		"scope":         "openid profile", // Remove email from scope
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", authURL, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var authResponse struct {
		IdToken   string `json:"id_token"`
		TokenType string `json:"token_type"`
		ExpiresIn int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", err
	}

	return authResponse.IdToken, nil
}
