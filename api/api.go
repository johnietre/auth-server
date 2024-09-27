package api

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/johnietre/auth-server/types"
	_ "github.com/mattn/go-sqlite3"
)

type DBConfig struct {
  UniqueEmail bool
  UniqueUsername bool
}

type DBClient struct {
  DB *sql.DB
}

func NewDBClient(db *sql.DB) *DBClient {
  return &DBClient{DB: db}
}

type InsertError struct {
  Err error
}

func (ie *InsertError) Error() string {
  if ie.Err != nil {
    return ie.Err.Error()
  }
  return ""
}

// NewUser creates a new user with the given password.
func (dbc *DBClient) NewUser(user *types.User, pwdHash string) error {
  res, err := dbc.DB.Exec(
    `INSERT INTO users(email,username,password_hash,firstname,lastname,other_info)
    VALUES (?,?,?,?,?,?)`,
    user.Email, user.Username, pwdHash,
    user.FirstName, user.LastName, user.OtherInfo,
  )
  // TODO
  if err != nil {
    return &InsertError{err}
  }
  id, err := res.LastInsertId()
  if err != nil {
    return err
  }
  user.Id = id
  return nil
}

var (
  // ErrNoUser is returned when a user cannot be found.
  ErrNoUser = errors.New("no user found")
)

// GetUser gets a single user matching the user info provided and populates the
// passed user with the info from the database.
func (dbc *DBClient) GetUser(userPtrs *types.UserPtrs) (*types.User, error) {
  clause := whereClauseFromUserPtrs(userPtrs)
  row := dbc.DB.QueryRow(
    `SELECT users(id,email,username,firstname,lastname,other_info) FROM users`+clause,
  )
  user := &types.User{}
  err := row.Scan(
    &user.Id, &user.Email, &user.Email, &user.Username,
    &user.FirstName, &user.LastName, &user.OtherInfo,
  )
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return nil, ErrNoUser
    } else {
      return nil, err
    }
  }
  return user, nil
}

// GetUsers gets all users matching the user info provided. If an error occurs
// while going through the result rows, iteration continues until the end, but
// the first error is returned along with the partial result.
func (dbc *DBClient) GetUsers(userPtrs *types.UserPtrs) ([]*types.User, error) {
  clause := whereClauseFromUserPtrs(userPtrs)
  rows, retErr := dbc.DB.Query(
    `SELECT users(id,email,username,firstname,lastname,other_info) FROM users`+clause,
  )
  if retErr != nil {
    if errors.Is(retErr, sql.ErrNoRows) {
      return nil, ErrNoUser
    }
    return nil, retErr
  }
  var users []*types.User
  for rows.Next() {
    user := &types.User{}
    err := rows.Scan(
      &user.Id, &user.Email, &user.Email, &user.Username,
      &user.FirstName, &user.LastName, &user.OtherInfo,
    )
    if err != nil {
      if retErr == nil {
        retErr = err
      }
    } else {
      users = append(users, user)
    }
  }
  return users, nil
}

// CheckPassword checks the provided password hash against what's stored in the
// database for a single user matching the user info provided.
func (dbc *DBClient) CheckPassword(user *types.UserPtrs, pwdHash string) (bool, error) {
  clause := whereClauseFromUserPtrs(user)
  row, hash := dbc.DB.QueryRow(`SELECT password_hash FROM users`+clause), ""
  if err := row.Scan(&hash); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return false, ErrNoUser
    } else {
      return false, err
    }
  }
  return CheckPassword(pwdHash, hash)
}

// EditUser edits a single user with the given info, setting the fields in
// newInfo to the appropriate values. The ID field cannot be changed.
func (dbc *DBClient) EditUser(userInfo, newInfo *types.UserPtrs) error {
  id := newInfo.Id
  newInfo.Id = nil
  // Dont include potential ID when setting new values
  parts := partsFromUserPtrs(newInfo)
  newInfo.Id = id
  if len(parts) == 0 {
    return nil
  }
  sets := strings.Join(parts, ",")
  clause := whereClauseFromUserPtrs(userInfo)
  _, err := dbc.DB.Exec(`UPDATE users SET `+sets+clause+` LIMIT 1`)
  return err
}

// EditUsers edits all users with the given info, setting the fields in
// newInfo to the appropriate values. The ID field cannot be changed.
func (dbc *DBClient) EditUsers(userInfo, newInfo *types.UserPtrs) error {
  id := newInfo.Id
  newInfo.Id = nil
  // Dont include potential ID when setting new values
  parts := partsFromUserPtrs(newInfo)
  newInfo.Id = id
  if len(parts) == 0 {
    return nil
  }
  sets := strings.Join(parts, ",")
  clause := whereClauseFromUserPtrs(userInfo)
  _, err := dbc.DB.Exec(`UPDATE users SET `+sets+clause)
  return err
}

// DeleteUser deletes a single user with information matching the provided
// user.
func (dbc *DBClient) DeleteUser(user *types.UserPtrs, clauses ...string) error {
  clause := whereClauseFromUserPtrs(user)
  if _, err := dbc.DB.Exec(`DELETE FROM users`+clause+` LIMIT 1`); err != nil {
    return err
  }
  return nil
}

// DeleteUser deletes all users with information matching the provided
// user.
func (dbc *DBClient) DeleteUsers(user *types.UserPtrs) error {
  clause := whereClauseFromUserPtrs(user)
  if _, err := dbc.DB.Exec(`DELETE FROM users`+clause); err != nil {
    return err
  }
  return nil
}

// Close closes the underlying DB connection.
func (dbc *DBClient) Close() error {
  return dbc.DB.Close()
}

func partsFromUserPtrs(user *types.UserPtrs) []string {
  clauses := []string{}
  // TODO: Handle better?
  if id := user.GetId(); id > 0 {
    clauses = append(clauses, fmt.Sprintf("id=%d", id))
  }
  if user.Email != nil {
    clauses = append(clauses, "email="+user.GetEmail())
  }
  if user.Username != nil {
    clauses = append(clauses, "username="+user.GetUsername())
  }
  if user.FirstName != nil {
    clauses = append(clauses, "first_name="+user.GetFirstName())
  }
  if user.LastName != nil {
    clauses = append(clauses, "last_name="+user.GetLastName())
  }
  if user.OtherInfo != nil {
    clauses = append(clauses, "other_info="+user.GetOtherInfo())
  }
  return clauses
}

// Returns a where clause if needed (prefixed by " WHERE ") or an empty string.
func whereClauseFromUserPtrs(user *types.UserPtrs) string {
  clauses := partsFromUserPtrs(user)
  if len(clauses) != 0 {
    return " WHERE "+strings.Join(clauses, " AND ")
  }
  return ""
}

const createTableStmt = `
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT %UNIQUE_EMAIL% NOT NULL,
  username TEXT %UNIQUE_USERNAME% NOT NULL,
  password_hash TEXT NOT NULL,
  firstname TEXT NOT NULL,
  lastname TEXT NOT NULL,
  other_info TEXT NOT NULL,
);
`
