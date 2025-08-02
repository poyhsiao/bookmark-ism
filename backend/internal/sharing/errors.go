package sharing

import "errors"

// Sharing service errors
var (
	ErrShareNotFound           = errors.New("share not found")
	ErrCollectionNotFound      = errors.New("collection not found")
	ErrInvalidShareToken       = errors.New("invalid share token")
	ErrShareExpired            = errors.New("share has expired")
	ErrShareInactive           = errors.New("share is inactive")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrInvalidCollectionID     = errors.New("invalid collection ID")
	ErrInvalidShareType        = errors.New("invalid share type")
	ErrInvalidPermission       = errors.New("invalid permission")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidName             = errors.New("invalid name")
	ErrCollaboratorExists      = errors.New("collaborator already exists")
	ErrCannotForkOwnCollection = errors.New("cannot fork own collection")
	ErrForkNotAllowed          = errors.New("fork not allowed for this collection")
	ErrInsufficientPermission  = errors.New("insufficient permission")
)
