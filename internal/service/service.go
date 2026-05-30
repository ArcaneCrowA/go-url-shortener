package service

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Shorten(site string) (string, error) {
	code, err := uuid.NewUUID()
	if err != nil {
		log.Printf("can't create uuid: %v", err)
		return "", errors.New("can't create uuid")
	}

	shortCode := code.String()[:6]
	if err = s.repo.InsertRecord(shortCode, site); err != nil {
		return "", err
	}

	return shortCode, nil
}

func (s *Service) Reroute(code string) (string, error) {
	site, err := s.repo.GetRecord(code)

	return site, err
}
