// Copyright 2025 James Ross
package rbacandtokens

import (
	"errors"
	"fmt"
)

// Authentication errors
var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrTokenNotYetValid  = errors.New("token is not yet valid")
	ErrInvalidSignature  = errors.New("invalid token signature")
	ErrMissingToken      = errors.New("missing authentication token")
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrRevokedToken      = errors.New("token has been revoked")
	ErrInvalidKeyID      = errors.New("invalid key ID")
	ErrKeyNotFound       = errors.New("signing key not found")
)

// Authorization errors
var (
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrInvalidRole             = errors.New("invalid role")
	ErrInvalidPermission       = errors.New("invalid permission")
	ErrAccessDenied            = errors.New("access denied")
	ErrResourceNotAllowed      = errors.New("resource access not allowed")
)

// Token management errors
var (
	ErrTokenGeneration = errors.New("failed to generate token")
	ErrKeyGeneration   = errors.New("failed to generate key pair")
	ErrKeyRotation     = errors.New("failed to rotate keys")
)

// AuthenticationError represents an authentication-related error
type AuthenticationError struct {
	Err     error
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error [%s]: %s", e.Code, e.Message)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(err error, code, message string, details map[string]interface{}) *AuthenticationError {
	return &AuthenticationError{
		Err:     err,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// AuthorizationError represents an authorization-related error
type AuthorizationError struct {
	Err       error
	Code      string
	Message   string
	Subject   string
	Resource  string
	Action    string
	RequiredPermissions []Permission
	UserPermissions     []Permission
	Details   map[string]interface{}
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error [%s]: %s for subject %s", e.Code, e.Message, e.Subject)
}

func (e *AuthorizationError) Unwrap() error {
	return e.Err
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(err error, code, message, subject, resource, action string, details map[string]interface{}) *AuthorizationError {
	return &AuthorizationError{
		Err:      err,
		Code:     code,
		Message:  message,
		Subject:  subject,
		Resource: resource,
		Action:   action,
		Details:  details,
	}
}

// TokenError represents a token-specific error
type TokenError struct {
	Err     error
	Code    string
	Message string
	TokenID string
	Subject string
	Details map[string]interface{}
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("token error [%s]: %s", e.Code, e.Message)
}

func (e *TokenError) Unwrap() error {
	return e.Err
}

// NewTokenError creates a new token error
func NewTokenError(err error, code, message, tokenID, subject string, details map[string]interface{}) *TokenError {
	return &TokenError{
		Err:     err,
		Code:    code,
		Message: message,
		TokenID: tokenID,
		Subject: subject,
		Details: details,
	}
}