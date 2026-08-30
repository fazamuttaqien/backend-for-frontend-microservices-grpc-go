package auth

import (
	"strings"
	"testing"
	"time"
)

func TestJWTIssueAndParse(t *testing.T){j,e:=NewJWT(strings.Repeat("s",32),"user-service",time.Hour);if e!=nil{t.Fatal(e)};tok,e:=j.Issue("user-1","user@example.com");if e!=nil{t.Fatal(e)};c,e:=j.Parse(tok);if e!=nil{t.Fatal(e)};if c.Subject!="user-1"||c.Email!="user@example.com"||c.Issuer!="user-service"{t.Fatal("claims mismatch")}}
func TestJWTRejectsTampering(t *testing.T){j,_:=NewJWT(strings.Repeat("s",32),"user-service",time.Hour);tok,_:=j.Issue("user-1","user@example.com");tok=strings.TrimSuffix(tok,"x")+"x";if _,e:=j.Parse(tok);e==nil{t.Fatal("tampered token accepted")}}
func TestJWTRejectsWrongSecret(t *testing.T){a,_:=NewJWT(strings.Repeat("a",32),"user-service",time.Hour);b,_:=NewJWT(strings.Repeat("b",32),"user-service",time.Hour);tok,_:=a.Issue("user-1","user@example.com");if _,e:=b.Parse(tok);e==nil{t.Fatal("token signed by another secret accepted")}}
func TestJWTRequiresStrongSecret(t *testing.T){if _,e:=NewJWT("short","user-service",time.Hour);e==nil{t.Fatal("weak secret accepted")}}
