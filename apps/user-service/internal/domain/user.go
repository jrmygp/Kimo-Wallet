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
)

// phoneNumberPattern matches E.164: a leading '+', a non-zero first digit,
// and 7-15 digits total (ITU-T E.164 max length).
var phoneNumberPattern = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

const maxFullNameLength = 100

// User is the identity record owned by the User Service.
type User struct {
	ID          string
	PhoneNumber string
	FullName    string
	CreatedAt   time.Time
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
