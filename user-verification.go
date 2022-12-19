package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type UserVerification interface {
	GenerateToken(request *VerficationRequest) (token string)

	VerifyToken(token string) (request *VerficationRequest, err error)
}

type VerficationRequest struct {
	guildId    string
	userId     string
	expiringAt int64
}

type MapUserVerification struct {
	tokens map[string]*VerficationRequest
}

func NewMapUserVerification() UserVerification {
	return &MapUserVerification{
		tokens: make(map[string]*VerficationRequest),
	}
}

func (s *MapUserVerification) GenerateToken(reqest *VerficationRequest) string {
	t := strconv.FormatUint(rand.Uint64(), 16) + strconv.FormatUint(rand.Uint64(), 16)
	s.tokens[t] = reqest
	return t
}

func (s *MapUserVerification) VerifyToken(token string) (request *VerficationRequest, err error) {
	request = s.tokens[token]
	if request == nil || request.expiringAt < time.Now().Unix() {
		return nil, fmt.Errorf("cannot find token")
	}
	delete(s.tokens, token)
	return request, nil
}
