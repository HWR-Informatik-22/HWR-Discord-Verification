package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	expiremap "github.com/nursik/go-expire-map"
)

type UserVerification interface {
	GenerateToken(request *VerficationRequest) (token string)

	VerifyToken(token string) (request *VerficationRequest, err error)

	Close()
}

type VerficationRequest struct {
	guildId    string
	userId     string
	expiringAt time.Time
}

type MapUserVerification struct {
	/* [string] *VerificationRequest */
	tokens *expiremap.ExpireMap
}

func NewMapUserVerification() UserVerification {
	return &MapUserVerification{
		tokens: expiremap.New(),
	}
}

func (s *MapUserVerification) GenerateToken(request *VerficationRequest) string {
	t := strconv.FormatUint(rand.Uint64(), 16) + strconv.FormatUint(rand.Uint64(), 16)
	fmt.Println("Token generated: "+t+". Which expire in", time.Until(request.expiringAt).Round(time.Second))
	s.tokens.Set(t, request, time.Until(request.expiringAt))
	return t
}

func (s *MapUserVerification) VerifyToken(token string) (request *VerficationRequest, err error) {
	value, ok := s.tokens.Get(token)
	if !ok {
		return nil, fmt.Errorf("cannot find token %s", token)
	}
	s.tokens.Delete(token)
	return value.(*VerficationRequest), nil
}

func (c *MapUserVerification) Close() {
	c.tokens.Close()
}
