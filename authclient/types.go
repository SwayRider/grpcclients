package authclient

import (
	"time"

	"github.com/swayrider/grpcclients/types"
	authv1 "github.com/swayrider/protos/auth/v1"
)

type User interface {
	UserId() string
	Email() string
	IsVerified() bool
	IsAdmin() bool
	AccountType() string
}

type UserCtor func(
	userId string,
	email string,
	isVerified bool,
	isAdmin bool,
	accountType string,
) User

type ServiceClient interface {
	ClientId() string
	Name() string
	Description() string
	Scopes() []string
}

type ServiceClientCtor func(
	clientId string,
	name string,
	description string,
	scopes ...string,
) ServiceClient

type Invite interface {
	Id() string
	Email() string
	CreatedAt() time.Time
	Registered() bool
}

type InviteCtor func(id string, email string, createdAt time.Time, registered bool) Invite

type WhoIsOneOf any

func WhoIs_Email(email string) WhoIsOneOf {
	return types.NewOneOf(
		email,
		func(v string) *authv1.WhoIsRequest_Email {
			return &authv1.WhoIsRequest_Email{
				Email: v,
			}
		},
	)
}

func WhoIs_UserId(userId string) WhoIsOneOf {
	return types.NewOneOf(
		userId,
		func(v string) *authv1.WhoIsRequest_UserId {
			return &authv1.WhoIsRequest_UserId{
				UserId: v,
			}
		},
	)
}
