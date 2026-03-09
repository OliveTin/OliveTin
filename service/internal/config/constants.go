package config

// ContentSecurityPolicyDefault is the default Content-Security-Policy header value
// when security headers are enabled and no custom policy is set.
const ContentSecurityPolicyDefault = "default-src 'self'; script-src 'self' 'unsafe-inline' https:; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https:; frame-ancestors 'none'; base-uri 'self'"
