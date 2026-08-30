package domain

import (
  "errors"
  "strings"
  "time"
)

var (
  ErrUserNotFound = errors.New("user not found")
  ErrEmailAlreadyExists = errors.New("email already exists")
  ErrInvalidUser = errors.New("invalid user")
)

type User struct { ID, Name, Email, PasswordHash string; CreatedAt, UpdatedAt time.Time }

func NewUser(id, name, email, passwordHash string, now time.Time) (*User,error) {
  name=strings.TrimSpace(name); email=strings.ToLower(strings.TrimSpace(email))
  if id=="" || name=="" || email=="" || passwordHash=="" { return nil,ErrInvalidUser }
  return &User{ID:id,Name:name,Email:email,PasswordHash:passwordHash,CreatedAt:now,UpdatedAt:now},nil
}
func (u *User) Update(name,email,passwordHash string, now time.Time) error {
  if strings.TrimSpace(name)=="" || strings.TrimSpace(email)=="" { return ErrInvalidUser }
  u.Name=strings.TrimSpace(name); u.Email=strings.ToLower(strings.TrimSpace(email))
  if passwordHash!="" { u.PasswordHash=passwordHash }; u.UpdatedAt=now; return nil
}
