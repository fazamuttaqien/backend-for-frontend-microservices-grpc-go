package userv1

import "google.golang.org/protobuf/reflect/protoreflect"

type User struct { Id string; Name string; Email string; CreatedAt string; UpdatedAt string }
type CreateUserRequest struct { Name string; Email string; Password string }
type CreateUserResponse struct { User *User }
type GetUserRequest struct { Id string }
type GetUserResponse struct { User *User }
type UpdateUserRequest struct { Id string; Name string; Email string; Password string }
type UpdateUserResponse struct { User *User }

func (*User) Reset() {}
func (*User) ProtoMessage() {}
func (*User) ProtoReflect() protoreflect.Message { return nil }
func (*CreateUserRequest) Reset() {}
func (*CreateUserRequest) ProtoMessage() {}
func (*CreateUserRequest) ProtoReflect() protoreflect.Message { return nil }
func (*CreateUserResponse) Reset() {}
func (*CreateUserResponse) ProtoMessage() {}
func (*CreateUserResponse) ProtoReflect() protoreflect.Message { return nil }
func (*GetUserRequest) Reset() {}
func (*GetUserRequest) ProtoMessage() {}
func (*GetUserRequest) ProtoReflect() protoreflect.Message { return nil }
func (*GetUserResponse) Reset() {}
func (*GetUserResponse) ProtoMessage() {}
func (*GetUserResponse) ProtoReflect() protoreflect.Message { return nil }
func (*UpdateUserRequest) Reset() {}
func (*UpdateUserRequest) ProtoMessage() {}
func (*UpdateUserRequest) ProtoReflect() protoreflect.Message { return nil }
func (*UpdateUserResponse) Reset() {}
func (*UpdateUserResponse) ProtoMessage() {}
func (*UpdateUserResponse) ProtoReflect() protoreflect.Message { return nil }
