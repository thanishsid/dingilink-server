package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	vd "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/thanishsid/mailgo"
	"github.com/thanishsid/tokenizer"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"

	"github.com/thanishsid/dingilink-server/internal/db"
	"github.com/thanishsid/dingilink-server/internal/model"
	"github.com/thanishsid/dingilink-server/internal/pkg/security"
	"github.com/thanishsid/dingilink-server/internal/types/apperror"
)

type UserService struct {
	DB                   db.DBQ
	Mail                 *mailgo.Client
	TokenConfig          tokenizer.Config
	EmailVerificationTTL time.Duration

	JwtAccessTokenTTL  time.Duration
	JwtRefreshTokenTTL time.Duration
}

type RegistrationInput struct {
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Image    *string `json:"image"`
	Bio      *string `json:"bio"`
	Password string  `json:"password"`

	EmailVerificationLink string `json:"emailVerificationLink"`
}

func (i RegistrationInput) Validate(usernameVd func(username string) error, emailVd func(email string) error) error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.Name, vd.Required.Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.Username,
			vd.Required.Error(apperror.INPUT_REQUIRED),
			vd.By(func(value interface{}) error {
				v := value.(string)
				return usernameVd(v)
			}),
		),
		vd.Field(&i.Email,
			vd.Required.Error(apperror.INPUT_REQUIRED),
			is.EmailFormat.Error(apperror.INPUT_INVALID),
			vd.By(func(value interface{}) error {
				v := value.(string)
				return emailVd(v)
			}),
		),

		//TODO - Add password strength validation
		vd.Field(&i.Password, vd.Required.Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.EmailVerificationLink, vd.Required.Error(apperror.INPUT_REQUIRED), is.URL.Error(apperror.INPUT_INVALID)),
	)
}

// User registration
func (s *UserService) Register(ctx context.Context, input RegistrationInput) error {
	if err := input.Validate(
		func(username string) error {
			exists, err := s.DB.CheckUsernameExists(ctx, username)
			if err != nil {
				return errors.New(apperror.INPUT_INVALID)
			}

			if exists {
				return errors.New(apperror.INPUT_DUPLICATE)
			}

			return nil
		},
		func(email string) error {
			exists, err := s.DB.CheckEmailExists(ctx, email)
			if err != nil {
				return errors.New(apperror.INPUT_INVALID)
			}

			if exists {
				return errors.New(apperror.INPUT_DUPLICATE)
			}

			return nil
		},
	); err != nil {
		return err
	}

	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert user to db
	userRes, err := tx.InsertUser(ctx, db.InsertUserParams{
		Username:     input.Username,
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: passwordHash,
		Image:        input.Image,
		Bio:          input.Bio,
	})
	if err != nil {
		return err
	}

	if err := tx.InsertUserRole(ctx, db.InsertUserRoleParams{
		UserID:   userRes.ID,
		RoleName: security.User.Name,
	}); err != nil {
		return err
	}

	token := uuid.New().String()

	_, err = tx.InsertEmailVerificationToken(ctx, db.InsertEmailVerificationTokenParams{
		UserID: userRes.ID,
		Token:  token,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(s.EmailVerificationTTL).UTC(),
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	email := userRes.Email

	verificationUrl, err := url.Parse(input.EmailVerificationLink)
	if err != nil {
		return err
	}

	q := verificationUrl.Query()

	q.Add("email", email)
	q.Add("token", token)

	verificationUrl.RawQuery = q.Encode()

	verificationLink := verificationUrl.String()

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	go func() {
		if err := s.Mail.SendMail(mailgo.SendMailParams{
			To:        []string{email},
			From:      "Dingilink",
			Subject:   "Dingilink Registration",
			PlainText: fmt.Sprintf("Welcome to Dingilink, please open the following link to verify your account. %s", verificationLink),
			TemaplateParams: &mailgo.SendMailTemplateParams{
				Name: "verification-link.html",
				Data: struct {
					Name     string
					Message  string
					Link     string
					Validity string
				}{
					Name:     userRes.Name,
					Message:  "Welcome to Dingilink, please open the following link to complete your registration",
					Link:     verificationLink,
					Validity: fmt.Sprintf("%d Minutes", s.EmailVerificationTTL/time.Minute),
				},
			},
		}); err != nil {
			log.Printf("failed to send email")
		}
	}()

	return nil
}

type EmailVerificationInput struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (i EmailVerificationInput) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.Email, vd.Required.Error(apperror.INPUT_REQUIRED), is.EmailFormat.Error(apperror.INPUT_INVALID)),
		vd.Field(&i.Token, vd.Required.Error(apperror.INPUT_REQUIRED)),
	)
}

// Verify user account email
func (s *UserService) VerifyEmail(ctx context.Context, input EmailVerificationInput) (*model.TokenPair, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tokenResult, err := tx.GetEmailVerificationToken(ctx, input.Token)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrInvalidEmailVerificationToken
	}
	if err != nil {
		return nil, err
	}

	if tokenResult.Email != input.Email {
		return nil, apperror.ErrUnexpected
	}

	if tokenResult.ExpiresAt.Time.Before(time.Now()) {
		return nil, apperror.ErrInvalidEmailVerificationToken
	}

	now := time.Now()

	if err := tx.UpdateUserEmailVerifiedAt(ctx, db.UpdateUserEmailVerifiedAtParams{
		EmailVerifiedAt: pgtype.Timestamptz{
			Time:  now.UTC(),
			Valid: true,
		},
		UserID: tokenResult.UserID,
	}); err != nil {
		return nil, err
	}

	if err := tx.DeleteEmailVerificationToken(ctx, tokenResult.ID); err != nil {
		return nil, err
	}

	user, err := tx.GetUser(ctx, tokenResult.UserID)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	accessTokenExpiry := currentTime.Add(s.JwtAccessTokenTTL)
	refreshTokenExpiry := currentTime.Add(s.JwtRefreshTokenTTL)

	tokenID := uuid.New()

	accessTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
		ID:        tokenID.String(),
	}

	refreshTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(refreshTokenExpiry),
		ID:        tokenID.String(),
	}

	accessToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, accessTokenClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, refreshTokenClaims)
	if err != nil {
		return nil, err
	}

	if err := tx.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		ID:     tokenID,
		UserID: user.ID,
		Token:  refreshToken,
		IssuedAt: pgtype.Timestamptz{
			Time:  currentTime.UTC(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshTokenExpiry.UTC(),
			Valid: true,
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type ResendEmailVerificationInput struct {
	Email                 string `json:"email"`
	EmailVerificationLink string `json:"emailVerificationLink"`
}

func (i ResendEmailVerificationInput) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.Email, vd.Required.Error(apperror.INPUT_REQUIRED), is.EmailFormat.Error(apperror.INPUT_INVALID)),
		vd.Field(&i.EmailVerificationLink, vd.Required.Error(apperror.INPUT_REQUIRED), is.URL.Error(apperror.INPUT_INVALID)),
	)
}

// Resend email verification
func (s *UserService) ResendEmailVerification(ctx context.Context, input ResendEmailVerificationInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userRes, err := tx.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return err
	}

	if userRes.EmailVerifiedAt.Valid {
		return apperror.ErrEmailAlreadyVerified
	}

	token := uuid.New().String()

	_, err = tx.InsertEmailVerificationToken(ctx, db.InsertEmailVerificationTokenParams{
		UserID: userRes.ID,
		Token:  token,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	email := userRes.Email

	verificationUrl, err := url.Parse(input.EmailVerificationLink)
	if err != nil {
		return err
	}

	q := verificationUrl.Query()

	q.Add("email", email)
	q.Add("token", token)

	verificationUrl.RawQuery = q.Encode()

	verificationLink := verificationUrl.String()

	go func() {
		if err := s.Mail.SendMail(mailgo.SendMailParams{
			To:        []string{email},
			From:      "Dingilink",
			Subject:   "Dingilink Registration",
			PlainText: fmt.Sprintf("Welcome to Dingilink, please open the following link to verify your account. %s", verificationLink),
			TemaplateParams: &mailgo.SendMailTemplateParams{
				Name: "verification-link.html",
				Data: struct {
					Name     string
					Message  string
					Link     string
					Validity string
				}{
					Name:     userRes.Name,
					Message:  "Welcome to Dingilink, please open the following link to complete your registration",
					Link:     verificationLink,
					Validity: "24 Hours",
				},
			},
		}); err != nil {
			log.Printf("failed to send email")
		}
	}()

	return nil
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i LoginInput) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.Email, vd.Required.Error(apperror.INPUT_REQUIRED), is.EmailFormat.Error(apperror.INPUT_INVALID)),
		vd.Field(&i.Password, vd.Required.Error(apperror.INPUT_REQUIRED)),
	)
}

// Login using email and password and obtain auth token pair.
func (s *UserService) Login(ctx context.Context, input LoginInput) (*model.TokenPair, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user, err := tx.GetUserByEmail(ctx, input.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrAccountNotFound
	} else if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	currentTime := time.Now()
	accessTokenExpiry := currentTime.Add(s.JwtAccessTokenTTL)
	refreshTokenExpiry := currentTime.Add(s.JwtRefreshTokenTTL)

	tokenID := uuid.New()

	accessTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
		ID:        tokenID.String(),
	}

	refreshTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(refreshTokenExpiry),
		ID:        tokenID.String(),
	}

	accessToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, accessTokenClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, refreshTokenClaims)
	if err != nil {
		return nil, err
	}

	if err := tx.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		ID:     tokenID,
		UserID: user.ID,
		Token:  refreshToken,
		IssuedAt: pgtype.Timestamptz{
			Time:  currentTime.UTC(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshTokenExpiry.UTC(),
			Valid: true,
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

type RefreshTokensInput struct {
	RefreshToken string `json:"refreshToken"`
}

func (i RefreshTokensInput) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.RefreshToken, vd.Required.Error(apperror.INPUT_REQUIRED)),
	)
}

// Refresh auth token pair by providing the refresh token.
func (s *UserService) RefreshTokens(ctx context.Context, input RefreshTokensInput) (*model.TokenPair, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var claims jwt.RegisteredClaims

	err := tokenizer.ParseToken(ctx, s.TokenConfig, input.RefreshToken, &claims)

	if errors.Is(err, tokenizer.ErrTokenExpired) {
		return nil, apperror.ErrTokenExpired
	} else if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, err
	}

	storedTokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	user, err := tx.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.DeletedAt.Valid {
		return nil, apperror.ErrAccountNotFound
	}

	storedRefreshToken, err := tx.GetRefreshToken(ctx, storedTokenID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.ErrTokenExpired
	} else if err != nil {
		return nil, err
	}

	currentTime := time.Now()

	if storedRefreshToken.ExpiresAt.Valid && storedRefreshToken.ExpiresAt.Time.Before(currentTime) {
		return nil, apperror.ErrTokenExpired
	}

	accessTokenExpiry := currentTime.Add(s.JwtAccessTokenTTL)
	refreshTokenExpiry := currentTime.Add(s.JwtRefreshTokenTTL)

	tokenID := uuid.New()

	accessTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
		ID:        tokenID.String(),
	}

	refreshTokenClaims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(refreshTokenExpiry),
		ID:        tokenID.String(),
	}

	accessToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, accessTokenClaims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokenizer.CreateToken(ctx, s.TokenConfig, refreshTokenClaims)
	if err != nil {
		return nil, err
	}

	if err := tx.InsertRefreshToken(ctx, db.InsertRefreshTokenParams{
		ID:     tokenID,
		UserID: user.ID,
		Token:  refreshToken,
		IssuedAt: pgtype.Timestamptz{
			Time:  currentTime.UTC(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  refreshTokenExpiry.UTC(),
			Valid: true,
		},
	}); err != nil {
		return nil, err
	}

	if err := tx.DeleteRefreshToken(ctx, storedTokenID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout clears current refresh token from the server
func (s *UserService) Logout(ctx context.Context) error {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return err
	}

	if err := s.DB.DeleteRefreshToken(ctx, userInfo.User.TokenID); err != nil {
		return err
	}

	return nil
}

// Logout from all devices clears all refresh tokens that belong to the current user from the server
func (s *UserService) LogoutFromAllDevices(ctx context.Context) error {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return err
	}

	if err := s.DB.DeleteRefreshTokensByUserID(ctx, userInfo.User.ID); err != nil {
		return err
	}

	return nil
}

// Get the currently logged in user account.
func (s *UserService) GetCurrentUser(ctx context.Context) (*model.User, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	user, err := s.DB.GetUser(ctx, userInfo.User.ID)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		Bio:         user.Bio,
		Image:       user.Image,
		Online:      user.Online,
		FriendCount: user.FriendCount,
	}, nil
}

type UpdateCurrentUserInput struct {
	Name             string  `json:"name"`
	Username         string  `json:"username"`
	Bio              *string `json:"bio"`
	Image            *string `json:"image"`
	ExistingPassword *string `json:"existingPassword"`
	NewPassword      *string `json:"newPassword"`
}

func (i UpdateCurrentUserInput) Validate() error {
	return vd.ValidateStruct(&i,
		vd.Field(&i.Name, vd.Required.Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.Username, vd.Required.Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.ExistingPassword, vd.Required.When(!vd.IsEmpty(i.NewPassword)).Error(apperror.INPUT_REQUIRED)),
		vd.Field(&i.NewPassword, vd.Required.When(!vd.IsEmpty(i.ExistingPassword)).Error(apperror.INPUT_REQUIRED)),
	)
}

func (s *UserService) UpdateCurrentUser(ctx context.Context, input UpdateCurrentUserInput) (*model.User, error) {
	userInfo, err := security.Authorize(ctx, security.User)
	if err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	updateParams := db.UpdateUserParams{
		UserID:   userInfo.User.ID,
		Username: input.Username,
		Name:     input.Name,
		Bio:      input.Bio,
		Image:    input.Image,
	}

	if input.ExistingPassword != nil && input.NewPassword != nil {
		user, err := s.DB.GetUser(ctx, userInfo.User.ID)
		if err != nil {
			return nil, err
		}

		if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(null.StringFromPtr(input.ExistingPassword).ValueOrZero())); err != nil {
			return nil, err
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(null.StringFromPtr(input.ExistingPassword).ValueOrZero()), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		updateParams.PasswordHash = passwordHash
	}

	if err := s.DB.UpdateUser(ctx, updateParams); err != nil {
		return nil, err
	}

	user, err := s.DB.GetUser(ctx, userInfo.User.ID)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		Bio:         user.Bio,
		Image:       user.Image,
		Online:      user.Online,
		FriendCount: user.FriendCount,
	}, nil
}

//TODO - ADD CREATE USER FUNCTION FOR ADMINS
