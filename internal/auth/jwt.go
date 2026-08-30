package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ( ErrInvalidToken=errors.New("invalid token"); ErrExpiredToken=errors.New("token expired") )
type Claims struct { Subject string `json:"sub"`; Email string `json:"email"`; Issuer string `json:"iss"`; Issued int64 `json:"iat"`; Expires int64 `json:"exp"` }
type contextKey string
const claimsKey contextKey="auth.claims"
type JWT struct { secret []byte; issuer string; ttl time.Duration; now func() time.Time }
func NewJWT(secret,issuer string,ttl time.Duration)(*JWT,error){if len(secret)<32{return nil,errors.New("JWT secret must be at least 32 bytes")};if strings.TrimSpace(issuer)==""||ttl<=0{return nil,errors.New("invalid JWT configuration")};return &JWT{[]byte(secret),issuer,ttl,time.Now},nil}
func (j *JWT) Issue(userID,email string)(string,error){if userID==""||email==""{return "",errors.New("user id and email are required")};n:=j.now().UTC();c:=Claims{Subject:userID,Email:email,Issuer:j.issuer,Issued:n.Unix(),Expires:n.Add(j.ttl).Unix()};h:=b64([]byte(`{"alg":"HS256","typ":"JWT"}`));pbytes,e:=json.Marshal(c);if e!=nil{return "",e};u:=h+"."+b64(pbytes);return u+"."+j.sign(u),nil}
func (j *JWT) Parse(token string)(*Claims,error){p:=strings.Split(token,".");if len(p)!=3||p[0]==""||p[1]==""||p[2]==""{return nil,ErrInvalidToken};hb,e:=base64.RawURLEncoding.DecodeString(p[0]);if e!=nil{return nil,ErrInvalidToken};var h struct{Alg string `json:"alg"`};if json.Unmarshal(hb,&h)!=nil||h.Alg!="HS256"{return nil,ErrInvalidToken};sig:=j.sign(p[0]+"."+p[1]);if subtle.ConstantTimeCompare([]byte(sig),[]byte(p[2]))!=1{return nil,ErrInvalidToken};pb,e:=base64.RawURLEncoding.DecodeString(p[1]);if e!=nil{return nil,ErrInvalidToken};var c Claims;if json.Unmarshal(pb,&c)!=nil||c.Subject==""||c.Email==""||c.Issuer!=j.issuer{return nil,ErrInvalidToken};if c.Expires<=j.now().Unix(){return nil,ErrExpiredToken};if c.Issued<=0||c.Issued>j.now().Add(time.Minute).Unix(){return nil,ErrInvalidToken};return &c,nil}
func(j *JWT)sign(v string)string{m:=hmac.New(sha256.New,j.secret);_,_=m.Write([]byte(v));return base64.RawURLEncoding.EncodeToString(m.Sum(nil))}
func b64(v []byte)string{return base64.RawURLEncoding.EncodeToString(v)}
func WithClaims(ctx context.Context,c *Claims)context.Context{return context.WithValue(ctx,claimsKey,c)}
func ClaimsFromContext(ctx context.Context)(*Claims,bool){c,ok:=ctx.Value(claimsKey).(*Claims);return c,ok}
