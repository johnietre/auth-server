package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/johnietre/auth-server/api"
	"github.com/johnietre/auth-server/types"
	utils "github.com/johnietre/utils/go"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
  UniqueEmail bool
  UniqueUsername bool
}

var (
  DbConfig Config

  logger = log.Default()
  db *sql.DB
)

func init() {
}

func SetLogger(l *log.Logger) {
  logger = l
}

func OpenDB(driverName, dataSourceName string) error {
  var err error
  db, err = sql.Open(driverName, dataSourceName)
  return err
}

func SetDB(database *sql.DB) {
  db = database
}

func CloseDB() error {
  return db.Close()
}

func Init() error {
  if db == nil {
    const (
      dn = "sqlite3"
      dsn = "file::memory:?cache=shared"
    )
    logger.Printf(
      "no database specified, opening using driverName=%s dataSourceName=%s",
      dn, dsn,
    )
    if err := OpenDB(dn, dsn); err != nil {
      return err
    }
  }
  stmt := createTableStmt
  if DbConfig.UniqueEmail {
    stmt = strings.Replace(stmt, "%UNIQUE_EMAIL%", "UNIQUE", 1)
  } else {
    stmt = strings.Replace(stmt, "%UNIQUE_EMAIL%", "", 1)
  }
  if DbConfig.UniqueUsername {
    stmt = strings.Replace(stmt, "%UNIQUE_USERNAME%", "UNIQUE", 1)
  } else {
    stmt = strings.Replace(stmt, "%UNIQUE_USERNAME%", "", 1)
  }
  db.Exec(stmt)
  return nil
}

func NewHandler() http.Handler {
  r := http.NewServeMux()
  return r
}

func createTokenHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    http.Error(w, "invalid method, expected POST", http.StatusMethodNotAllowed)
    return
  }
  req := types.CreateTokenRequest{}
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    if utils.IsUnmarshalError(err) {
      http.Error(w, err.Error(), http.StatusBadRequest)
    } else {
      logger.Printf("error decoding JSON: %v", err)
      httpISE(w)
    }
    return
  }
  if req.ShouldHash {
    pwd, err := api.GenerateHash(req.Password)
    if err != nil {
      if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
        http.Error(w, "password too long", http.StatusBadRequest)
      } else {
        logger.Printf("error hashing password: %v", err)
        httpISE(w)
      }
      return
    }
    req.Password = pwd
  }
  user, clauses := req.User, []string{}
  // TODO: Handle better?
  if user.Id > 0 {
    clauses = append(clauses, fmt.Sprintf("id=%d", user.Id))
  }
  if user.Email != "" {
    clauses = append(clauses, "email="+user.Email)
  }
  if user.Username != "" {
    clauses = append(clauses, "username="+user.Username)
  }
  clause := ""
  if len(clauses) != 0 {
    clause = " WHERE "+strings.Join(clauses, " AND ")
  }
  row, hash := db.QueryRow(`SELECT password_hash FROM users`+clause), ""
  if err := row.Scan(&hash); err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      http.Error(w, "no user found", http.StatusNotFound)
      return
    } else {
      logger.Printf("error querying user: %v", err)
      httpISE(w)
    }
  }
  if ok, err := api.CheckPassword(req.Password, hash); err != nil {
    logger.Printf("error checking password: %v", err)
    httpISE(w)
  } else if !ok {
    http.Error(w, "incorrect password", http.StatusUnauthorized)
  } else {
    // TODO
  }
}

func newUserHandler(w http.ResponseWriter, r *http.Request) {
  req := types.NewUserRequest{}
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    if utils.IsUnmarshalError(err) {
      http.Error(w, err.Error(), http.StatusBadRequest)
    } else {
      logger.Printf("error decoding JSON: %v", err)
      httpISE(w)
    }
    return
  }
  if req.ShouldHash {
    pwd, err := api.GenerateHash(req.Password)
    if err != nil {
      if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
        http.Error(w, "password too long", http.StatusBadRequest)
      } else {
        logger.Printf("error hashing password: %v", err)
        httpISE(w)
      }
      return
    }
    req.Password = pwd
  }
  user := req.User
  res, err := db.Exec(
    `INSERT INTO users(email,username,password_hash,firstname,lastname,other_info)
    VALUES (?,?,?,?,?,?)`,
    user.Email, user.Username, req.Password,
    user.Firstname, user.Lastname, user.OtherInfo,
  )
  if err != nil {
    logger.Printf("error inserting user: %v", err)
    httpISE(w)
    return
  }
  id, err := res.LastInsertId()
  if err != nil {
    logger.Printf("error getting new user ID: %v", err)
    httpISE(w)
    return
  }
  user.Id = id
  if err := json.NewEncoder(w).Encode(user); err != nil {
    if utils.IsMarshalError(err) {
      logger.Printf("error marshaling user: %v", err)
    }
    // TODO: log other errors?
  }
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
}

func isUnmarshalErr(err error) bool {
  ute, se := &json.UnmarshalTypeError{}, &json.SyntaxError{}
  return errors.As(err, &ute) || errors.As(err, &se)
}

const createTableStmt = `
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT %UNIQUE_EMAIL%,
  username TEXT %UNIQUE_USERNAME%,
  password_hash TEXT,
  firstname TEXT,
  lastname TEXT,
  other_info TEXT
);
`

func httpISE(w http.ResponseWriter) {
  http.Error(w, "internal server error", http.StatusInternalServerError)
}
