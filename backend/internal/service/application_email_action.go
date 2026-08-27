package service

import (
	"crypto/sha256"
	"net/url"
	"strconv"
	"strings"
	"time"

	"serverdock/internal/dto"

	"github.com/golang-jwt/jwt/v5"
)

const (
	emailActionAudience = "serverdock-email-application-action"
	emailActionIssuer   = "serverdock"
	emailActionTokenTTL = 7 * 24 * time.Hour
	emailActionEndpoint = "api/applications/public/email-action"
)

type emailActionClaims struct {
	ApplicationID uint   `json:"application_id"`
	Action        string `json:"action"`
	jwt.RegisteredClaims
}

type emailActionLinks struct {
	ApproveURL string
	RejectURL  string
	IgnoreURL  string
}

func (s *ApplicationService) HandleEmailAction(token string) (*dto.ApplicationResponse, error) {
	applicationID, action, err := s.parseEmailActionToken(token)
	if err != nil {
		return nil, err
	}

	switch action {
	case "approve":
		return s.Approve(applicationID)
	case "reject":
		return s.Reject(applicationID)
	case "ignore":
		return s.Ignore(applicationID)
	default:
		return nil, ErrInvalidEmailAction
	}
}

func (s *ApplicationService) newEmailActionLinks(applicationID uint) emailActionLinks {
	baseURL := normalizePublicURL(s.configService.Get("public_url"))
	if baseURL == "" {
		baseURL = normalizePublicURL(s.fallbackPublicURL)
	}
	if baseURL == "" || strings.TrimSpace(s.emailActionSecret) == "" {
		return emailActionLinks{}
	}

	endpoint, err := url.JoinPath(baseURL, emailActionEndpoint)
	if err != nil {
		return emailActionLinks{}
	}
	buildURL := func(action string) string {
		token, tokenErr := s.createEmailActionToken(applicationID, action)
		if tokenErr != nil {
			return ""
		}
		return endpoint + "?action=" + action + "#token=" + url.QueryEscape(token)
	}

	return emailActionLinks{
		ApproveURL: buildURL("approve"),
		RejectURL:  buildURL("reject"),
		IgnoreURL:  buildURL("ignore"),
	}
}

func (s *ApplicationService) createEmailActionToken(applicationID uint, action string) (string, error) {
	if applicationID == 0 || !isApplicationAction(action) || strings.TrimSpace(s.emailActionSecret) == "" {
		return "", ErrInvalidEmailAction
	}

	now := s.now().UTC()
	claims := emailActionClaims{
		ApplicationID: applicationID,
		Action:        action,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{emailActionAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(emailActionTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    emailActionIssuer,
			Subject:   strconv.FormatUint(uint64(applicationID), 10),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.emailActionSigningKey())
}

func (s *ApplicationService) parseEmailActionToken(tokenString string) (uint, string, error) {
	if strings.TrimSpace(tokenString) == "" || strings.TrimSpace(s.emailActionSecret) == "" {
		return 0, "", ErrInvalidEmailAction
	}

	claims := &emailActionClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidEmailAction
			}
			return s.emailActionSigningKey(), nil
		},
		jwt.WithAudience(emailActionAudience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(emailActionIssuer),
		jwt.WithTimeFunc(s.now),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid || claims.ApplicationID == 0 || !isApplicationAction(claims.Action) {
		return 0, "", ErrInvalidEmailAction
	}
	if claims.Subject != strconv.FormatUint(uint64(claims.ApplicationID), 10) {
		return 0, "", ErrInvalidEmailAction
	}
	return claims.ApplicationID, claims.Action, nil
}

func (s *ApplicationService) emailActionSigningKey() []byte {
	sum := sha256.Sum256([]byte("serverdock/email-action/v1\x00" + s.emailActionSecret))
	return sum[:]
}

func isApplicationAction(action string) bool {
	return action == "approve" || action == "reject" || action == "ignore"
}

func normalizePublicURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}
