package api

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// GenerateHash generates a hash for a password with the default cost.
func GenerateHash(pwd string) (string, error) {
  hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
  return string(hash), err
}

// CheckPassword checks a password against a hash, returning (true, nil) if the
// password is a match. Returns (false, nil) if the error returned from
// comparing the hash is bcrypt.ErrMismatchHashAndPassword or
// bcrypt.ErrorPasswordTooLong. In any other cases, (false, err) is returned.
func CheckPassword(pwd, hash string) (bool, error) {
  err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
  if err == nil {
    return true, nil
  } else if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrPasswordTooLong) {
    return false, nil
  }
  return false, err
}
