package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/rabbits"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/apis/service"
	"go.uber.org/zap"
)

type Handler struct {
	consumer rabbits.IConsumer
	service  *service.Service
	logger   *zap.Logger

	stopCtx context.Context
	errChan chan error
}

func NewHandler(consumer rabbits.IConsumer, service *service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		consumer: consumer,
		service:  service,
		logger:   logger,
	}
}

func (h *Handler) Start() {
	ctx := context.Background()
	h.errChan = make(chan error, 1)
	h.stopCtx = ctx

	go func() {
		for err := range h.errChan {
			_ = err
			h.Stop()
		}
	}()

	_ = consumeMessages(ctx, h.service.SendConfirmUserEmail, h.consumer, rabbits.ConfirmUserEmailQueueKey, h.logger, h.errChan)

	_ = consumeMessages(ctx, h.service.SendSuccessConfirmUserEmail, h.consumer, rabbits.SuccessConfirmUserEmailQueueKey, h.logger, h.errChan)
}

func (h *Handler) Stop() {
	h.stopCtx.Done()
	err := h.stopCtx.Err()
	if err != nil && errors.Is(err, context.Canceled) {
		h.logger.Info("Context canceled", zap.Error(err), zap.Any("context", h.stopCtx))
	} else if err != nil {
		h.logger.Info("Context error", zap.Error(err), zap.Any("context", h.stopCtx))
	}
}

func consumeMessages[T any](ctx context.Context, nextFunc func(*T) error, consumer rabbits.IConsumer, queueName string, logger *zap.Logger, errChan chan error) error {
	msgs, err := consumer.Consume(ctx, queueName)
	if err != nil {
		logger.Info("Consumer error", zap.Error(err), zap.Any("context", ctx))
		errChan <- err
		return err
	}

	logger.Info("Consumer started", zap.String("queue", queueName))

	go func() {
		for {
			for msg := range msgs {
				data, err := unmarshal[T](msg.Body)
				if err != nil {
					logger.Error("Failed to unmarshal message body", zap.Error(err))
				}

				err = nextFunc(data)
				if err != nil {
					logger.Error("Failed to send data", zap.Error(err))
					errChan <- err
				}
			}
		}
	}()

	return nil
}

func unmarshal[T any](data []byte) (*T, error) {
	v := new(T)

	err := json.Unmarshal(data, v)
	if err != nil {
		return v, err
	}

	return v, nil
}
