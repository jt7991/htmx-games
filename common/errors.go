package common

import "errors"

var UsernameTakenError = errors.New("Username is already taken")

var UnknownError = errors.New("An unknown error occurred")
