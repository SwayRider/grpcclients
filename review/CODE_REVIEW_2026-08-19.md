# Code Review — 2026-08-19

Security-focused review of the full current codebase of `grpcclients` (not diff-based). No prior `review/CODE_REVIEW_*.md` exists in this repo, so all findings below are new. See [`Docs/REVIEW.md`](../../Docs/REVIEW.md) for how findings in this file are tracked and marked fixed.

Scope covered: `client.go`, `types/types.go`, `internal/client/client.go` (shared base), and all six package `client.go`/`config.go` files (auth, health, mail, region, router, search). No test files exist. No logging calls exist anywhere in the module — errors are returned to callers, never printed, so there's no token/PII log-leakage risk within this package itself. No hardcoded credentials, no unsafe deserialization, and dependency versions (`google.golang.org/grpc v1.80.0`, `protobuf v1.36.11`) look current.

### 1. Plaintext gRPC transport for all inter-service traffic

`internal/client/client.go:27` — every client connects via `grpc.WithTransportCredentials(insecure.NewCredentials())`, unconditionally, with no TLS/mTLS option. The README (line 157) documents this as intentional, deferring TLS to the service mesh/load balancer. Failure scenario: if any service (or a debug/staging deployment) is reachable outside the intended mesh boundary — a misconfigured network policy, a container exposed on a shared subnet, cloud metadata/VPC peering misconfig — all inter-service traffic, including bearer JWTs, refresh tokens, service-client secrets (`GetToken`, `CreateServiceClient` responses), and passwords in `Register`/`ChangePassword`/`ResetPassword` requests, travels unencrypted and is trivially sniffable. Given this library carries credentials for every service-to-service call in the platform, the blast radius of a network-isolation failure is large. Recommend at minimum making mTLS configurable (even if disabled by default) so it isn't a code change to enable when the trust boundary changes. Severity: Medium.

### 2. Authentication relies entirely on JWT/network isolation, no service-identity/mTLS layer

No shared-secret or mTLS-based service authentication exists — auth is purely "bearer token in metadata" (user JWT or client-credentials token from `GetToken`), riding over the plaintext connection from finding #1. Combined with #1, a compromised pod/container on the internal network can both eavesdrop on and directly forge calls to any service without needing to steal a token first, since nothing else verifies the caller's identity at the transport layer. This is standard "trust the network" design and consistent with the README's stated model, but is flagged as it lacks defense-in-depth against lateral movement. Severity: Medium.

### 3. No retry policy or circuit breaker; single shared connection per client

`internal/client/client.go` and every package's `client.go`: only a per-RPC deadline (5s default, 60s for router) bounds calls — there is no retry-with-backoff, no circuit breaker, and each logical client holds exactly one `grpc.ClientConn`. A slow-but-not-dead downstream (e.g. authservice under load) causes every dependent caller to block up to the full deadline on every request with no fallback, which can cascade under load (thundering herd once the deadline trips and all callers retry at the application layer simultaneously, with no jitter/backoff built in). Not exploitable directly, but a plausible internal DoS amplifier. Severity: Low.

### 4. Unsynchronized read of the connection pointer after `CheckConnection` returns

Pattern repeated in every package (e.g. `authclient/client.go:58-72`, `mailclient/client.go:34-48`): `CheckConnection()` takes `mux` and may reassign `c.Client`/`c.healthClient`, but the subsequent `c.Impl()` call in each RPC method happens after the lock is released. Under concurrent use of a single client instance across goroutines (common for a shared package-level client), one goroutine's reconnect can race with another's read of the connection pointer — Go's race detector would flag this. Worst case is a spurious RPC failure on a connection mid-swap, not memory corruption, but it undermines the intended "callers never manage connection state" guarantee under load. Severity: Low.

### 5. `regionclient` README claims "no auth required" but every method forwards a bearer token

`README.md:12` vs. `regionclient/client.go` (`SearchPoint`, `SearchBox`, `SearchRadius`, `FindCrossingLocations`, etc.) — all forward `Authorization: Bearer <token>`. Either the README is stale or regionservice doesn't actually enforce the token it's sent (auth theater). Cross-check: regionservice's own review (`regionservice/review/CODE_REVIEW_2026-08.md`, finding #1) confirms the `region:query` scope check is enforced server-side, so this is a stale README, not an enforcement gap. Worth a doc fix. Severity: Info.
