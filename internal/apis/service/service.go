package service

import (
	"fmt"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/apis/mailer"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/models"
	"go.uber.org/zap"
)

type Service struct {
	mailer *mailer.GoMailer
	logger *zap.Logger
}

func NewService(mailer *mailer.GoMailer, logger *zap.Logger) *Service {
	return &Service{
		mailer: mailer,
		logger: logger,
	}
}

func (s *Service) SendConfirmUserEmail(data *models.ConfirmEmailMail) error {
	err := s.mailer.SendConfirmMessage(data)
	if err != nil {
		s.logger.Error("send confirm code email err", zap.Error(err))
		return fmt.Errorf("send confirm code email err: %v", err)
	}

	return nil
}

func (s *Service) SendSuccessConfirmUserEmail(data *models.SuccessConfirmEmailMail) error {
	err := s.mailer.SendSuccessConfirmMessage(data)
	if err != nil {
		s.logger.Error("send success confirm code email err", zap.Error(err))
		return fmt.Errorf("send success confirm code email err: %v", err)
	}

	return nil
}
