// Package carp handles requests in a reverse-proxy manner and supports these use-cases against the target dogu:
//   - general browser requests with deliberate accounts
//   - browser logout requests with CAS accounts
//   - general REST requests with CAS accounts
//   - general REST requests with dogu-internal accounts
//
// In general all failed (that includes REST) requests will lead to a redirect response towards CAS. Also failed
// REST requests are subject to request limiting because CAS must not be concerned with dogu-internal accounts - which
// may also lead to undefined throttle behavior on the CAS side.
package carp
