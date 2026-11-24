// Package http documents the shared HTTP subpackages extracted from Shibuya.
// Import the individual packages rather than this documentation-only package.
//
// # Module
//
// The module is github.com/reqfleet/lib. It has no dependency on the Shibuya
// application or engine-sdk. Run GOWORK=off go test -race -count=1 ./... from
// the repository root.
//
// # Routing and middleware
//
// route.Router assembles method-aware net/http ServeMux routes and nested path
// prefixes. Registration mutates the supplied routes; complete it before serving.
// Route.MakePattern replaces Shibuya's misspelled MakePttern method.
//
// auth.AuthRequiredWithToken applies strict, case-sensitive Bearer authentication
// to every request. Missing/malformed credentials return 400 and mismatches
// return 403. Tokens are compared in constant time for equal-length values.
// An empty configured secret denies all requests with 503 without invoking the
// protected handler. Public endpoints such as /healthz are application routing
// policy and must be served outside this middleware.
//
// Logging uses chi's configurable middleware.DefaultLogger. Compose exclusions
// outside the logger, without Shibuya's unused handler arguments:
//
//	handler := auth.ExcludePathsFromLogger(
//		map[string]struct{}{"/healthz": {}}, nil,
//	)(auth.RequestLoggerWithoutPaths(router.Mux()))
//
// Exclusions match URL.Path exactly or by a configured substring, not the query.
// Using the logger alone logs every request. Configure the logger before wrapping
// handlers. Logging is opt-in; routing and authentication do not install it.
//
// # Clients, servers, and tokens
//
// client supplies the three-second and custom-timeout HTTP client constructors.
// Callers continue to own transports, trust roots, and idle-connection cleanup.
// server.StartServer serves HTTP or HTTPS and drains requests on context
// cancellation. Unlike Shibuya's application helper, it does not own process
// signals, wait for a signal after external shutdown, or terminate the process.
// Applications may pass a signal.NotifyContext; nil disables automatic shutdown.
//
// auth/token only parses Bearer credentials. JWT signing/verification, key-file
// conventions, session cookies, CSRF policy, and token lifetimes remain in
// Shibuya; they are not dependencies of this library. redirect validates
// same-origin absolute-path redirect references.
//
// Application-specific account/session authorization, database access, and API
// middleware remain in Shibuya. This extraction does not migrate its callers.
package http
