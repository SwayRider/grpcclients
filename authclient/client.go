package authclient

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"google.golang.org/grpc"
	"github.com/swayrider/grpcclients"
	"github.com/swayrider/grpcclients/internal/client"
	"github.com/swayrider/grpcclients/types"
	authv1 "github.com/swayrider/protos/auth/v1"
	healthv1 "github.com/swayrider/protos/health/v1"
)

type Client struct {
	*client.Client[authv1.AuthServiceClient]
	healthClient   healthv1.HealthServiceClient
	mux            *sync.Mutex
	getHostAndPort grpcclients.GetHostAndPort
}

func New(getHostAndPort grpcclients.GetHostAndPort) (grpcclients.Client, error) {
	c := &Client{
		mux:            &sync.Mutex{},
		getHostAndPort: getHostAndPort,
	}

	err := c.newConnection()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) CheckConnection() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.Client == nil {
		return c.newConnection()
	}

	err := c.Ping()
	if err != nil {
		c.Close()
		return c.newConnection()
	}
	return nil
}

func (c *Client) newConnection() error {
	host, port := c.getHostAndPort()
	var healthClient healthv1.HealthServiceClient
	clnt, err := client.New(
		host, port,
		func(conn *grpc.ClientConn) authv1.AuthServiceClient {
			healthClient = healthv1.NewHealthServiceClient(conn)
			return authv1.NewAuthServiceClient(conn)
		},
		Deadline(),
	)
	if err != nil {
		return err
	}

	c.Client = clnt
	c.healthClient = healthClient
	return nil
}

func (c Client) Ping() error {
	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err := c.healthClient.Ping(ctx, &healthv1.PingRequest{})
	return err
}

func (c Client) CheckPasswordStrength(
	password string,
) (
	isStrong bool,
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().CheckPasswordStrength(
		ctx,
		&authv1.CheckPasswordStrengthRequest{
			Password: password,
		},
	)
	if err != nil {
		return
	}

	isStrong = res.IsStrong
	message = res.Message
	return
}

func (c Client) CreateAdmin(
	accessToken string,
	email, password string,
) (
	userId string,
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().CreateAdmin(
		ctx,
		&authv1.CreateAdminRequest{
			Email:    email,
			Password: password,
		},
	)
	if err != nil {
		return
	}

	userId = res.UserId
	message = res.Message
	return
}

func (c Client) Register(
	email, password, verificationUrl string,
) (
	userId string,
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().Register(
		ctx,
		&authv1.RegisterRequest{
			Email:           email,
			Password:        password,
			VerificationUrl: verificationUrl,
		},
	)
	if err != nil {
		return
	}

	userId = res.UserId
	message = res.Message
	return
}

func (c Client) VerifyEmail(
	email, verificationUrl string,
) (
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err = c.Impl().VerifyEmail(
		ctx,
		&authv1.VerifyEmailRequest{
			Email:           email,
			VerificationUrl: verificationUrl,
		},
	)
	if err != nil {
		return
	}

	return
}

func (c Client) Login(
	email, password string,
	rememberMe bool,
) (
	accessToken string,
	refreshToken string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().Login(
		ctx,
		&authv1.LoginRequest{
			Email:      email,
			Password:   password,
			RememberMe: rememberMe,
		},
	)
	if err != nil {
		return
	}

	accessToken = res.AccessToken
	refreshToken = res.RefreshToken
	return
}

func (c Client) Logout(
	refreshToken string,
) (
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err = c.Impl().Logout(
		ctx,
		&authv1.LogoutRequest{
			RefreshToken: refreshToken,
		},
	)
	if err != nil {
		return
	}
	return
}

func (c Client) GetToken(
	clientId, clientSecret string,
	scopes []string,
) (
	accessToken string,
	grantedScopes []string,
	validUntil time.Time,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().GetToken(
		ctx,
		&authv1.GetTokenRequest{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			Scopes:       scopes,
		},
	)
	if err != nil {
		return
	}

	accessToken = res.AccessToken
	grantedScopes = res.Scopes
	validUntil = res.ValidUntil.AsTime()
	return
}

func (c Client) Refresh(
	refreshToken string,
	rememberMe bool,
) (
	newAccessToken string,
	newRefreshToken string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().Refresh(
		ctx,
		&authv1.RefreshRequest{
			RefreshToken: refreshToken,
			RememberMe:   rememberMe,
		},
	)
	if err != nil {
		return
	}

	newAccessToken = res.AccessToken
	newRefreshToken = res.RefreshToken
	return
}

func (c Client) ChangePassword(
	accessToken string,
	oldPassword, newPassword string,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().ChangePassword(
		ctx,
		&authv1.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		},
	)
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) ChangeAccountType(
	accessToken string,
	userId, accountType string,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().ChangeAccountType(
		ctx,
		&authv1.ChangeAccountTypeRequest{
			UserId:      userId,
			AccountType: accountType,
		},
	)
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) WhoAmI(
	accessToken string,
	userCtor UserCtor,
) (
	user User,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().WhoAmI(ctx, &authv1.WhoAmIRequest{})
	if err != nil {
		return
	}

	user = userCtor(
		res.UserId,
		res.Email,
		res.IsVerified,
		res.IsAdmin,
		res.AccountType,
	)
	return
}

func (c Client) WhoIs(
	accessToken string,
	userIdOrEmail WhoIsOneOf,
	userCtor UserCtor,
) (
	user User,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	req := &authv1.WhoIsRequest{}

	switch v := userIdOrEmail.(type) {
	case types.PbOneOf[*authv1.WhoIsRequest_UserId]:
		req.WhoisOneof = v.PbStruct()
	case types.PbOneOf[*authv1.WhoIsRequest_Email]:
		req.WhoisOneof = v.PbStruct()
	}

	res, err := c.Impl().WhoIs(ctx, req)
	if err != nil {
		return
	}

	user = userCtor(
		res.UserId,
		res.Email,
		res.IsVerified,
		res.IsAdmin,
		res.AccountType,
	)
	return
}

func (c Client) CreateVerificationToken(
	accessToken string,
) (
	userId string,
	token string,
	validUnitl time.Time,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().CreateVerificationToken(ctx, &authv1.CreateVerificationTokenRequest{})
	if err != nil {
		return
	}

	userId = res.UserId
	token = res.Token
	validUnitl = res.ValidUntil.AsTime()
	return
}

func (c Client) CheckVerificationToken(
	userId string,
	token string,
) (
	valid bool,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().CheckVerificationToken(ctx, &authv1.CheckVerificationTokenRequest{
		UserId: userId,
		Token:  token,
	})
	if err != nil {
		return
	}

	valid = res.IsValid
	return
}

func (c Client) RequestPasswordReset(
	email string,
	resetUrl string,
) (
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	_, err = c.Impl().RequestPasswordReset(ctx, &authv1.RequestPasswordResetRequest{
		Email:    email,
		ResetUrl: resetUrl,
	})
	if err != nil {
		return
	}

	return
}

func (c Client) ResetPassword(
	userId string,
	token string,
	newPasword string,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().ResetPassword(ctx, &authv1.ResetPasswordRequest{
		UserId:      userId,
		Token:       token,
		NewPassword: newPasword,
	})
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) PublicKeys() (
	publicKeys []string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.Context(context.Background())
	defer cancel()

	res, err := c.Impl().PublicKeys(ctx, &authv1.PublicKeysRequest{})
	if err != nil {
		return
	}

	publicKeys = make([]string, 0, len(res.Keys))
	for _, k := range res.Keys {
		pubKeyByes, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}

		publicKeys = append(publicKeys, string(pubKeyByes))
	}
	return
}

func (c Client) CreateServiceClient(
	accessToken string,
	name string,
	description string,
	scopes []string,
) (
	clientId string,
	clientSecret string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().CreateServiceClient(ctx, &authv1.CreateServiceClientRequest{
		Name:        name,
		Description: description,
		Scopes:      scopes,
	})
	if err != nil {
		return
	}

	clientId = res.ClientId
	clientSecret = res.ClientSecret
	return
}

func (c Client) DeleteServiceClient(
	accessToken string,
	clientId string,
) (
	message string,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().DeleteServiceClient(ctx, &authv1.DeleteServiceClientRequest{
		ClientId: clientId,
	})
	if err != nil {
		return
	}

	message = res.Message
	return
}

func (c Client) ListServiceClients(
	accessToken string,
	page int,
	pageSize int,
	serviceClientCtor ServiceClientCtor,
) (
	clients []ServiceClient,
	err error,
) {
	if err = c.CheckConnection(); err != nil {
		return
	}

	ctx, cancel := c.AuthorizedContext(context.Background(), accessToken)
	defer cancel()

	res, err := c.Impl().ListServiceClients(ctx, &authv1.ListServiceClientsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return
	}

	clients = make([]ServiceClient, len(res.Clients))
	for i, client := range res.Clients {
		clients[i] = serviceClientCtor(
			client.ClientId,
			client.Name,
			client.Description,
			client.Scopes...,
		)
	}
	return
}
