package domain

import "errors"

// Sentinel errors shared by the domain, store and server layers.
var (
	// ErrNotFound means the requested record does not exist.
	ErrNotFound = errors.New("forever: not found")
	// ErrAmbiguous means an identifier resolved to more than one catalog row.
	ErrAmbiguous = errors.New("forever: ambiguous reference")
	// ErrUnsupportedClass means a class key is not part of the pinned catalog.
	ErrUnsupportedClass = errors.New("forever: unsupported class")
)
