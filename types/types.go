package types

// User holds user data.
type User struct {
  Id int64 `json:"id,omitempty"`
  Email string `json:"email,omitempty"`
  Username string `json:"username,omitempty"`
  FirstName string `json:"first_name,omitempty"`
  LastName string `json:"last_name,omitempty"`
  OtherInfo string `json:"other_info,omitempty"`
}

// ToUserPtrs creates a new UserPtrs with from the User. The returned UserPtrs
// and the originating User are not interdependent (through pointers).
func (u User) ToUserPtrs() *UserPtrs {
  return &UserPtrs{
    Id: &u.Id,
    Email: &u.Email,
    Username: &u.Username,
    FirstName: &u.FirstName,
    LastName: &u.LastName,
    OtherInfo: &u.OtherInfo,
  }
}

// UserPtrs is used for when wanting to use optional fields, for example, when
// wanting to know what to include in a query.
type UserPtrs struct {
  Id *int64 `json:"id,omitempty"`
  Email *string `json:"email,omitempty"`
  Username *string `json:"username,omitempty"`
  FirstName *string `json:"first_name,omitempty"`
  LastName *string `json:"last_name,omitempty"`
  OtherInfo *string `json:"other_info,omitempty"`
}

// GetId returns the Id if not nil, otherwise, 0.
func (up *UserPtrs) GetId() int64 {
  if up.Id != nil {
    return *up.Id
  }
  return 0
}

// GetEmail returns the Email if not nil, otherwise, an empty string.
func (up *UserPtrs) GetEmail() string {
  if up.Email != nil {
    return *up.Email
  }
  return ""
}

// GetUsername returns the Username if not nil, otherwise, an empty string.
func (up *UserPtrs) GetUsername() string {
  if up.Username != nil {
    return *up.Username
  }
  return ""
}

// GetFirstName returns the FirstName if not nil, otherwise, an empty string.
func (up *UserPtrs) GetFirstName() string {
  if up.FirstName != nil {
    return *up.FirstName
  }
  return ""
}

// GetLastName returns the LastName if not nil, otherwise, an empty string.
func (up *UserPtrs) GetLastName() string {
  if up.LastName != nil {
    return *up.LastName
  }
  return ""
}

// GetOtherInfo returns the OtherInfo if not nil, otherwise, an empty string.
func (up *UserPtrs) GetOtherInfo() string {
  if up.OtherInfo != nil {
    return *up.OtherInfo
  }
  return ""
}

// ToUser creates a new User. The returned User is not interdependent with the
// originating UserPtrs (through pointers).
func (up *UserPtrs) ToUser() *User {
  return &User{
    Id: up.GetId(),
    Email: up.GetEmail(),
    Username: up.GetUsername(),
    FirstName: up.GetFirstName(),
    LastName: up.GetLastName(),
    OtherInfo: up.GetOtherInfo(),
  }
}

// CreateTokenRequest is the datatype expected when requesting to create a new
// token.
type CreateTokenRequest struct {
  User User `json:"user"`
  Password string `json:"password,omitempty"`
  ShouldHash bool `json:"shouldHash,omitempty"`
}

// NewUserRequest is the datatype expected when requesting to create a new
// user.
type NewUserRequest struct {
  User User `json:"user"`
  Password string `json:"password,omitempty"`
  ShouldHash bool `json:"shouldHash,omitempty"`
}

func NewT[T any](t T) *T {
  return &t
}
