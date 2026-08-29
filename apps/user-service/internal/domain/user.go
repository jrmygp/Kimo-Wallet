package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalidPhoneNumber = errors.New("phone number must be in E.164 format, e.g. +6281234567890")
	ErrInvalidFullName    = errors.New("full name must be between 1 and 100 characters")
	ErrPhoneNumberTaken   = errors.New("phone number is already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidKimoID      = errors.New("kimo id must be 12 uppercase alphanumeric characters")

	// ErrKimoIDCollision means the randomly generated KimoID collided with
	// an existing one on insert. Never surfaced past the service layer —
	// Register catches it and retries with a freshly generated KimoID
	// (see internal/service/user_service.go) — a real caller never needs
	// to know a retry happened underneath.
	ErrKimoIDCollision = errors.New("kimo id already exists")
)

// phoneNumberPattern matches E.164: a leading '+', a non-zero first digit,
// and 7-15 digits total (ITU-T E.164 max length).
var phoneNumberPattern = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// kimoIDPattern matches the shape idgen.NewKimoID produces: exactly 12
// uppercase alphanumeric characters. Rejecting a malformed KimoID here,
// before it ever reaches the database, is what turns "not a KimoID at
// all" into a 400 instead of a generic no-match falling through as a 500
// — see docs/agent-logs/2026-08-28.md Entry 4 (the same fix, previously
// applied when this field was the internal UUID `id`).
var kimoIDPattern = regexp.MustCompile(`^[A-Z0-9]{12}$`)

const maxFullNameLength = 100

// User is the identity record owned by the User Service.
type User struct {
	ID          string
	PhoneNumber string
	FullName    string
	CreatedAt   time.Time
	// ProfilePicture is nil when the user has none set.
	ProfilePicture *string
	// KimoID is the public-facing identifier — see kimoIDPattern below.
	KimoID string
}

// RegisterInput is the validated, ready-to-persist form of a registration request.
type RegisterInput struct {
	PhoneNumber string
	FullName    string
}

type LoginInput struct {
	PhoneNumber string
}

// NewRegisterInput validates raw registration fields and returns a RegisterInput,
// or the first validation error encountered. Validation here is a UX/shape check;
// it is not the authority on whether the phone number is already taken.
func NewRegisterInput(phoneNumber, fullName string) (RegisterInput, error) {
	phoneNumber = strings.TrimSpace(phoneNumber)
	if !phoneNumberPattern.MatchString(phoneNumber) {
		return RegisterInput{}, ErrInvalidPhoneNumber
	}

	fullName = strings.TrimSpace(fullName)
	nameLength := utf8.RuneCountInString(fullName)
	if nameLength == 0 || nameLength > maxFullNameLength {
		return RegisterInput{}, ErrInvalidFullName
	}

	return RegisterInput{PhoneNumber: phoneNumber, FullName: fullName}, nil
}

func NewLoginInput(phoneNumber string) (LoginInput, error) {
	phoneNumber = strings.TrimSpace(phoneNumber)
	if !phoneNumberPattern.MatchString(phoneNumber) {
		return LoginInput{}, ErrInvalidPhoneNumber
	}

	return LoginInput{PhoneNumber: phoneNumber}, nil
}

// NewGetUserByKimoIDInput validates that kimoID is shaped like a KimoID
// and returns it trimmed and upper-cased, or ErrInvalidKimoID. Upper-casing
// before validating/querying means a search is case-insensitive even
// though every stored KimoID is canonically uppercase — a typed-in lookup
// shouldn't fail just because someone typed it in lowercase.
//
// This is a shape check only — same as NewLoginInput's — not an authority
// on whether that KimoID actually exists; ErrUserNotFound covers
// "well-formed but doesn't exist".
func NewGetUserByKimoIDInput(kimoID string) (string, error) {
	kimoID = strings.ToUpper(strings.TrimSpace(kimoID))
	if !kimoIDPattern.MatchString(kimoID) {
		return "", ErrInvalidKimoID
	}

	return kimoID, nil
}
