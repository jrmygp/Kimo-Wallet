// Package domain holds this service's core types, validation, and sentinel
// errors — kept free of gRPC/ORM/HTTP concerns so it can be tested and read
// without any of those (see docs/CLAUDE.md §5.4).
//
// TEMPLATE NOTE: "Widget" stands in for this service's real entity. Rename
// the file, the type, every identifier below, and widgets_table in the
// migration to match — see apps/service-template/README.md for the full
// checklist.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrWidgetNotFound = errors.New("widget not found")
	ErrInvalidID      = errors.New("id must be a valid uuid")
	ErrInvalidName    = errors.New("name must be between 1 and 100 characters")
)

// idPattern matches a v4 UUID as produced by internal/idgen.NewV4.
var idPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

const maxNameLength = 100

// Widget is the record this service owns. Replace its fields with the real
// entity's — this shape (an id, one describing field, a status, a
// creation time) is deliberately the smallest useful example, not a
// pattern to preserve.
type Widget struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
}

// CreateWidgetInput is the validated, ready-to-persist form of a create
// request. Validation here is a shape/UX check — it is not the authority
// on business rules that need the database (uniqueness, ownership, etc.);
// those are enforced in the repository/service layers instead. See
// docs/CLAUDE.md §5.1 "validate at the boundary".
type CreateWidgetInput struct {
	Name string
}

func NewCreateWidgetInput(name string) (CreateWidgetInput, error) {
	name = strings.TrimSpace(name)
	nameLength := utf8.RuneCountInString(name)
	if nameLength == 0 || nameLength > maxNameLength {
		return CreateWidgetInput{}, ErrInvalidName
	}

	return CreateWidgetInput{Name: name}, nil
}

// NewGetWidgetByIDInput validates that id is shaped like a UUID and
// returns it trimmed, or ErrInvalidID. This is a shape check only —
// ErrWidgetNotFound covers "well-formed but doesn't exist".
func NewGetWidgetByIDInput(id string) (string, error) {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return "", ErrInvalidID
	}

	return id, nil
}
