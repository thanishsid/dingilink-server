package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/thanishsid/mailgo"
	"github.com/thanishsid/tokenizer"
	"golang.org/x/sync/errgroup"

	"github.com/thanishsid/dingilink-server/api"
	"github.com/thanishsid/dingilink-server/asset"
	"github.com/thanishsid/dingilink-server/internal/config"
	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
	"github.com/thanishsid/dingilink-server/internal/pkg/messaging"
	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/services"
)

func main() {
	// Context with cancellation on SIGINT and SIGKILL
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	tokenConfig, err := tokenizer.NewHMAC(jwt.SigningMethodHS256, []byte(cfg.JwtSecretKey))
	if err != nil {
		log.Fatal(err)
	}

	pgconn, err := db.ConnectPool(mainCtx, cfg.DBConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	pg := db.NewDBQuerier(pgconn)

	awsConf := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, ""),
		Region:      cfg.AwsRegion,
	}

	s3Client := s3.NewFromConfig(awsConf)

	mailClient, err := mailgo.NewClient(mailgo.DialerConfig{
		Host:      cfg.SmtpHost,
		Port:      cfg.SmtpPort,
		Username:  cfg.SmtpEmail,
		Password:  cfg.SmtpPassword,
		Templates: asset.MailTemplates,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := security.InitSecurity(mainCtx, pg); err != nil {
		log.Fatal(err)
	}

	messageEventchannelManager, err := messaging.NewChannelManager[*model.MessageEvent](cfg.NatsUrl)
	if err != nil {
		log.Fatal(err)
	}

	uploadService := &services.UploadService{
		S3Client:  s3Client,
		S3Bucket:  cfg.S3Bucket,
		S3BaseDir: "test",
	}

	userService := &services.UserService{
		DB:                   pg,
		Mail:                 mailClient,
		TokenConfig:          tokenConfig,
		EmailVerificationTTL: time.Minute * 60,
		JwtAccessTokenTTL:    time.Hour * 24 * 2,
		JwtRefreshTokenTTL:   time.Hour * 24 * 60,
	}

	messageService := &services.MessageService{
		DB: pg,
		CH: messageEventchannelManager,
	}

	h := api.NewHandler(
		&api.HandlerConfig{
			UploadService:  uploadService,
			UserService:    userService,
			MessageService: messageService,
		},
	)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: h,
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	g, gCtx := errgroup.WithContext(mainCtx)

	// Start server in separate goroutine
	g.Go(func() error {
		fmt.Printf("\nServer running on port %s !!\n", cfg.ServerPort)
		return srv.ListenAndServe()
	})

	// Listen for context cancellation in seprate goroutine and call server shutdown.
	g.Go(func() error {
		<-gCtx.Done()
		return srv.Shutdown(context.Background())
	})

	// Wait Indefinitelty
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
