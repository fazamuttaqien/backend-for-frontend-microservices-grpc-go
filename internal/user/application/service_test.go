package application_test

import (
	"context"
	"testing"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/user/domain"
	"golang.org/x/crypto/bcrypt"
)

type repo struct{ users map[string]*domain.User }
func(r *repo)Create(_ context.Context,u *domain.User)error{if r.users==nil{r.users=map[string]*domain.User{}};r.users[u.ID]=u;return nil}
func(r *repo)GetByID(_ context.Context,id string)(*domain.User,error){u,ok:=r.users[id];if !ok{return nil,domain.ErrUserNotFound};return u,nil}
func(r *repo)GetByEmail(_ context.Context,email string)(*domain.User,error){for _,u:=range r.users{if u.Email==email{return u,nil}};return nil,domain.ErrUserNotFound}
func(r *repo)Update(_ context.Context,u *domain.User)error{r.users[u.ID]=u;return nil}
func TestCreateHashesPassword(t *testing.T){r:=&repo{};s:=application.NewUserService(r);u,e:=s.Create(context.Background(),"Faza","FAZA@EXAMPLE.COM","secret");if e!=nil{t.Fatal(e)};if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash),[]byte("secret"))!=nil{t.Fatal("password hash does not match")};if u.PasswordHash=="secret"{t.Fatal("plaintext password stored")}}
func TestAuthenticate(t *testing.T){r:=&repo{};s:=application.NewUserService(r);u,_:=s.Create(context.Background(),"Faza","faza@example.com","secret");if _,e:=s.Authenticate(context.Background(),"FAZA@EXAMPLE.COM","secret");e!=nil{t.Fatal(e)};if _,e:=s.Authenticate(context.Background(),u.Email,"wrong");e!=domain.ErrInvalidCredentials{t.Fatal("wrong password should fail with invalid credentials")};if _,e:=s.Authenticate(context.Background(),"missing@example.com","secret");e!=domain.ErrInvalidCredentials{t.Fatal("unknown email should fail with invalid credentials")}}
func TestGetAndUpdate(t *testing.T){r:=&repo{};s:=application.NewUserService(r);u,e:=s.Create(context.Background(),"Old","old@example.com","secret");if e!=nil{t.Fatal(e)};got,e:=s.Get(context.Background(),u.ID);if e!=nil||got.ID!=u.ID{t.Fatal(e)};updated,e:=s.Update(context.Background(),u.ID,"New","new@example.com","");if e!=nil{t.Fatal(e)};if updated.Name!="New"||updated.Email!="new@example.com"{t.Fatal("update failed")};if updated.PasswordHash!=u.PasswordHash{t.Fatal("password hash changed unexpectedly")}}
