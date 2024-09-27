package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/johnietre/auth-server/types"
	utils "github.com/johnietre/utils/go"
)

type Non200Error struct {
  Code int
  Body []byte
  Other error
}

func (e *Non200Error) Error() string {
  if e.Other == nil {
    return fmt.Sprintf("%d status: %s", e.Code, e.Body)
  }
  return fmt.Sprintf("%d status: %s (other error: %v)", e.Code, e.Body, e.Other)
}

type HttpClient struct {
  Addr string
  AutoAuth bool
  Token *utils.AValue[string]
  HttpClient *http.Client
}

func NewHttpClient(addr string) *HttpClient {
  return &HttpClient{
    Addr: addr,
    AutoAuth: true,
    Token: utils.NewAValue(""),
    HttpClient: &http.Client{},
  }
}

func (hc *HttpClient) GetToken(
  user *types.User, password string, shouldHash bool,
) (string, error) {
  jsonBytes, err := json.Marshal(user)
  if err != nil {
    return "", err
  }
  body := bytes.NewReader(jsonBytes)
  req, err := http.NewRequest("POST", path.Join(hc.Addr, "/token"), body)
  if err != nil {
    // NOTE: Shouldn't occur, do something else?
    return "", err
  }
  resp, err := hc.HttpClient.Do(req)
  if err != nil {
    return "", err
  }
  respBody, err := io.ReadAll(resp.Body)
  resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", &Non200Error{Code: resp.StatusCode, Body: respBody, Other: err}
  }
  return string(respBody), err
}

func (hc *HttpClient) NewUser(
  user *types.User, password string, shouldHash bool,
) error {
  return hc.newUser(user, password, shouldHash, false)
}

func (hc *HttpClient) newUser(
  user *types.User, password string, shouldHash bool, ran bool,
) error {
  jsonBytes, err := json.Marshal(user)
  if err != nil {
    return err
  }
  body := bytes.NewReader(jsonBytes)
  req, err := http.NewRequest("POST", path.Join(hc.Addr, "/users"), body)
  if err != nil {
    // NOTE: Shouldn't occur, do something else?
    return err
  }
  resp, err := hc.HttpClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    // TODO: Is this necessary
    if resp.StatusCode == http.StatusUnauthorized && hc.AutoAuth && !ran {
      tok, err := hc.GetToken(user, password, shouldHash)
      if err == nil {
        hc.Token.Store(tok)
        err = hc.newUser(user, password, shouldHash, true)
      }
      return err
    }
    e := &Non200Error{Code: resp.StatusCode}
    e.Body, e.Other = io.ReadAll(resp.Body)
    return e
  }
  newUser := types.User{}
  if err := json.NewDecoder(resp.Body).Decode(&newUser); err != nil {
    return err
  }
  *user = newUser
  return nil
}
