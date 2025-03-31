package server

import (
	"context"
	"fmt"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/abstractions"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/consul"
	apperrors "github.com/Brain-Wave-Ecosystem/go-common/pkg/error"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/log"
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/rabbits"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/apis/handler"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/apis/mailer"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/apis/service"
	"github.com/Brain-Wave-Ecosystem/mail-service/internal/config"
	"github.com/DavidMovas/gopherbox/pkg/closer"
	"github.com/NawafSwe/gomailer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"time"
)

var _ abstractions.Server = (*Server)(nil)

type Server struct {
	grpcServer *grpc.Server
	handler    *handler.Handler
	consul     *consul.Consul
	logger     *log.Logger
	cfg        *config.Config
	closer     *closer.Closer
}

func NewServer(_ context.Context, cfg *config.Config) (*Server, error) {
	cl := closer.NewCloser()

	logger, err := log.NewLogger(cfg.Local, cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	cl.Push(logger.Stop)

	consulManager, err := consul.NewConsul(cfg.ConsulURL, cfg.Name, cfg.Address, cfg.GRPCPort, logger.Zap())
	if err != nil {
		logger.Zap().Error("error initializing consul manager", zap.Error(err))
		return nil, fmt.Errorf("error initializing consul manager: %w", err)
	}

	cl.Push(consulManager.Stop)

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor())

	healthServer := health.NewServer()
	healthServer.SetServingStatus(fmt.Sprintf("%s-%d", cfg.Name, cfg.GRPCPort), grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	cl.PushNE(healthServer.Shutdown)

	goMailer, err := initMailer()
	if err != nil {
		return nil, err
	}

	consumer, err := rabbits.NewConsumer(cfg.Rabbit.URL, rabbits.JSON,
		rabbits.WithExchange(rabbits.ExchangeKey, rabbits.ExchangeDirect, true, false, nil),
		rabbits.WithQueueAndBind(rabbits.ConfirmUserEmailQueueKey, rabbits.ConfirmUserEmailKey, rabbits.ExchangeKey, true, false, nil),
		rabbits.WithQueueAndBind(rabbits.SuccessConfirmUserEmailQueueKey, rabbits.SuccessConfirmUserEmailKey, rabbits.ExchangeKey, true, false, nil),
	)
	if err != nil {
		logger.Zap().Error("Failed to create rabbits consumer", zap.Error(err))
		return nil, err
	}

	cl.PushIO(consumer)

	s := service.NewService(goMailer, logger.Zap())
	h := handler.NewHandler(consumer, s, logger.Zap())

	cl.PushNE(h.Stop)

	return &Server{
		grpcServer: grpcServer,
		handler:    h,
		consul:     consulManager,
		logger:     logger,
		cfg:        cfg,
		closer:     cl,
	}, nil
}

func (s *Server) Start() error {
	z := s.logger.Zap()

	z.Info("Starting server", zap.String("name", s.cfg.Name), zap.Int("port", s.cfg.GRPCPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GRPCPort))
	if err != nil {
		z.Error("Failed to start listener", zap.String("name", s.cfg.Name), zap.Int("port", s.cfg.GRPCPort), zap.Error(err))
		return err
	}

	s.closer.PushIO(lis)

	err = s.consul.RegisterService()
	if err != nil {
		z.Error("Failed to register service in consul registry", zap.String("name", s.cfg.Name), zap.Error(err))
		return err
	}

	return s.grpcServer.Serve(lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	z := s.logger.Zap()

	z.Info("Shutting down server", zap.String("name", s.cfg.Name))

	s.grpcServer.GracefulStop()

	<-ctx.Done()
	s.grpcServer.Stop()

	return s.closer.Close(ctx)
}

func initMailer() (*mailer.GoMailer, error) {
	mailerCfg := &mailer.Config{
		Host:   "mail_mailhog",
		Port:   1025,
		Sender: "brain-wave@gmail.com",
		Options: []gomailer.Options{
			gomailer.WithDialTimeout(time.Second * 10),
			gomailer.WithLocalName("localhost"),
		},
	}

	goMailer, err := mailer.NewGoMailer(mailerCfg)
	if err != nil {
		return nil, apperrors.InternalWithoutStackTrace(err)
	}

	return goMailer, nil
}
