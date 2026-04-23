# grpcclients

Go client library for all SwayRider gRPC services. Each sub-package wraps one service from `github.com/swayrider/protos` and exposes a clean, proto-free API to callers.

## Packages

| Package | Service | Auth required |
|---|---|---|
| `authclient` | AuthService | some methods |
| `healthclient` | HealthService | no |
| `mailclient` | MailService | some methods |
| `regionclient` | RegionService | no |
| `routerclient` | RouterService | yes |
| `searchclient` | SearchService | yes |

## Basic usage

Every client is created with a `GetHostAndPort` function — a `func() (string, int)` that is called each time the client (re)connects. This allows the address to come from a service-discovery mechanism or environment variable rather than being hardcoded at startup.

```go
import "github.com/swayrider/grpcclients/regionclient"

client, err := regionclient.New(func() (string, int) {
    return os.Getenv("REGION_HOST"), 9000
})
if err != nil {
    log.Fatal(err)
}
defer client.Close()

regions, err := client.SearchPoint(regionclient.Coordinate{Lat: 50.85, Lon: 4.35}, false)
```

All clients satisfy the `grpcclients.Client` interface:

```go
type Client interface {
    CheckConnection() error
    Close() error
}
```

`CheckConnection` pings the service and transparently reconnects if the connection is stale. Every method calls it before making an RPC, so callers do not need to manage connection state.

## Constructor injection

Methods that return domain objects (users, routes, search results…) never instantiate the caller's concrete types. Instead they accept a **constructor function** (`UserCtor`, `RouteCtor`, `SearchResultCtor`, …) and call it to build the return value. This keeps proto-generated types out of the caller's domain model.

```go
import "github.com/swayrider/grpcclients/authclient"

// Define your own type that satisfies authclient.User
type AppUser struct { id, email string; verified, admin bool; kind string }
func (u *AppUser) UserId() string      { return u.id }
func (u *AppUser) Email() string       { return u.email }
func (u *AppUser) IsVerified() bool    { return u.verified }
func (u *AppUser) IsAdmin() bool       { return u.admin }
func (u *AppUser) AccountType() string { return u.kind }

// Pass a constructor — authclient calls it and returns your type
user, err := authClient.WhoAmI(accessToken, func(id, email string, verified, admin bool, kind string) authclient.User {
    return &AppUser{id, email, verified, admin, kind}
})
```

The same pattern is used by `routerclient` (`RouteCtor`, `RouteStepCtor`) and `searchclient` (`SearchResultCtor`).

## Authenticated calls

Methods that require a user JWT accept `accessToken string` as their first argument. The client sets the `Authorization: Bearer <token>` metadata header automatically — callers never touch gRPC metadata.

```go
message, err := mailClient.Send(accessToken, mail)
```

Services with "internal" variants (mail) accept no token and are intended for server-to-server calls:

```go
message, err := mailClient.SendInternal(mail)
```

## Deadlines

Each package exposes a package-level deadline (default 5 s, 60 s for routerclient) that is applied to every RPC context:

```go
routerclient.SetDeadline(90 * time.Second)
d := routerclient.Deadline()
```

## authclient extras

### WhoIs oneof

`WhoIs` accepts either an email or a user ID. Use the provided helpers to build the oneof value without importing proto types:

```go
user, err := authClient.WhoIs(accessToken, authclient.WhoIs_Email("alice@example.com"), userCtor)
user, err := authClient.WhoIs(accessToken, authclient.WhoIs_UserId("usr_123"), userCtor)
```

### PublicKeyFetcher

For JWT verification, `authclient.PublicKeyFetcher` is a background goroutine that fetches the service's public keys, sends them on a channel, and then refreshes every hour:

```go
keysChan := make(chan []string, 1)
go authclient.PublicKeyFetcher(ctx, authClient, keysChan)
keys := <-keysChan // blocks until first fetch succeeds
```

## Keeping clients in sync with protos

The proto definitions live in `github.com/swayrider/protos`. When they change, check each affected client against this checklist:

**New service added**
- Create a new `<name>client/` package with `client.go`, `config.go`, and `types.go`.
- Follow the structure of any existing client — `regionclient` is the simplest reference.
- Add an entry to the table at the top of this file.

**New RPC added to an existing service**
- Add a method to the client's `client.go`.
- Use `c.Context()` for unauthenticated calls, `c.AuthorizedContext()` for calls that require a Bearer token.
- If the response includes domain objects, add an interface + constructor type in `types.go` rather than returning raw proto types.

**New field on an existing message**
- Add the corresponding field to the Go struct or interface in `types.go`.
- Wire it through in `client.go` where the proto struct is constructed.
- For structs, adding a field is non-breaking as long as all callers use named field syntax (standard Go practice).
- For interface return types, adding a method is a breaking change — prefer a new method on the client if the existing interface is already in use externally.

**RPC removed from a proto**
- Remove the method from the client, but first grep the entire repo for callers.
- Remove any types in `types.go` that are now unused — again, check for external usage first.

**Verify after any change**

```sh
# from this directory
go build ./...

# from any service that imports grpcclients
go build ./...
```

## Internal structure

```
grpcclients/
├── client.go               # GetHostAndPort type + Client interface
├── types/
│   └── types.go            # PbOneOf — generic helper for proto oneof fields
├── internal/
│   └── client/
│       └── client.go       # Client[T] — generic base used by all concrete clients
├── authclient/             # AuthService
├── healthclient/           # HealthService
├── mailclient/             # MailService
├── regionclient/           # RegionService
├── routerclient/           # RouterService
└── searchclient/           # SearchService
```

`internal/client.Client[T]` handles the gRPC connection, deadline wrapping, and JWT metadata injection. Concrete clients embed it and only need to implement service-specific method mappings.

All connections use insecure transport credentials (`grpc.WithTransportCredentials(insecure.NewCredentials())`). TLS termination is expected to be handled by the service mesh or load balancer in front of each service.
