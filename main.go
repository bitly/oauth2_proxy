package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/mreiferson/go-options"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flagSet := flag.NewFlagSet("oauth2_proxy", flag.ExitOnError)

	emailDomains := StringArray{}
	upstreams := StringArray{}
	skipAuthRegex := StringArray{}
	googleGroups := StringArray{}
	httpAllowedHosts := StringArray{}
	httpHostsProxyHeaders := StringArray{}

	config := flagSet.String("config", "", "path to config file")
	showVersion := flagSet.Bool("version", false, "print version string")

	flagSet.String("http-address", "127.0.0.1:4180", "[http://]<addr>:<port> or unix://<path> to listen on for HTTP clients")
	flagSet.String("https-address", ":443", "<addr>:<port> to listen on for HTTPS clients")
	flagSet.String("tls-cert", "", "path to certificate file")
	flagSet.String("tls-key", "", "path to private key file")
	flagSet.String("redirect-url", "", "the OAuth Redirect URL. ie: \"https://internalapp.yourcompany.com/oauth2/callback\"")
	flagSet.Bool("set-xauthrequest", false, "set X-Auth-Request-User and X-Auth-Request-Email response headers (useful in Nginx auth_request mode)")
	flagSet.Var(&upstreams, "upstream", "the http url(s) of the upstream endpoint or file:// paths for static files. Routing is based on the path")
	flagSet.Bool("pass-basic-auth", true, "pass HTTP Basic Auth, X-Forwarded-User and X-Forwarded-Email information to upstream")
	flagSet.Bool("pass-user-headers", true, "pass X-Forwarded-User and X-Forwarded-Email information to upstream")
	flagSet.String("basic-auth-password", "", "the password to set when passing the HTTP Basic Auth header")
	flagSet.Bool("pass-access-token", false, "pass OAuth access_token to upstream via X-Forwarded-Access-Token header")
	flagSet.Bool("pass-host-header", true, "pass the request Host Header to upstream")
	flagSet.Var(&skipAuthRegex, "skip-auth-regex", "bypass authentication for requests path's that match (may be given multiple times)")
	flagSet.Bool("skip-provider-button", false, "will skip sign-in-page to directly reach the next step: oauth/start")
	flagSet.Bool("skip-auth-preflight", false, "will skip authentication for OPTIONS requests")
	flagSet.Bool("ssl-insecure-skip-verify", false, "skip validation of certificates presented when using HTTPS")

	flagSet.Var(&emailDomains, "email-domain", "authenticate emails with the specified domain (may be given multiple times). Use * to authenticate any email")
	flagSet.String("azure-tenant", "common", "go to a tenant-specific or common (tenant-independent) endpoint.")
	flagSet.String("github-org", "", "restrict logins to members of this organisation")
	flagSet.String("github-team", "", "restrict logins to members of this team")
	flagSet.Var(&googleGroups, "google-group", "restrict logins to members of this google group (may be given multiple times).")
	flagSet.String("google-admin-email", "", "the google admin to impersonate for api calls")
	flagSet.String("google-service-account-json", "", "the path to the service account json credentials")
	flagSet.String("client-id", "", "the OAuth Client ID: ie: \"123456.apps.googleusercontent.com\"")
	flagSet.String("client-secret", "", "the OAuth Client Secret")
	flagSet.String("authenticated-emails-file", "", "authenticate against emails via file (one per line)")
	flagSet.String("htpasswd-file", "", "additionally authenticate against a htpasswd file. Entries must be created with \"htpasswd -s\" for SHA encryption or \"htpasswd -B\" for bcrypt encryption")
	flagSet.Bool("display-htpasswd-form", true, "display username / password login form if an htpasswd file is provided")
	flagSet.String("custom-templates-dir", "", "path to custom html templates")
	flagSet.String("footer", "", "custom footer string. Use \"-\" to disable default footer.")
	flagSet.String("proxy-prefix", "/oauth2", "the url root path that this proxy should be nested under (e.g. /<oauth2>/sign_in)")

	flagSet.String("cookie-name", "_oauth2_proxy", "the name of the cookie that the oauth_proxy creates")
	flagSet.String("cookie-secret", "", "the seed string for secure cookies (optionally base64 encoded)")
	flagSet.String("cookie-domain", "", "an optional cookie domain to force cookies to (ie: .yourcompany.com)*")
	flagSet.Duration("cookie-expire", time.Duration(168)*time.Hour, "expire timeframe for cookie")
	flagSet.Duration("cookie-refresh", time.Duration(0), "refresh the cookie after this duration; 0 to disable")
	flagSet.Bool("cookie-secure", true, "set secure (HTTPS) cookie flag")
	flagSet.Bool("cookie-httponly", true, "set HttpOnly cookie flag")

	flagSet.Bool("request-logging", true, "Log requests to stdout")
	flagSet.String("request-logging-format", defaultRequestLoggingFormat, "Template for log lines")

	flagSet.String("provider", "google", "OAuth provider")
	flagSet.String("oidc-issuer-url", "", "OpenID Connect issuer URL (ie: https://accounts.google.com)")
	flagSet.String("login-url", "", "Authentication endpoint")
	flagSet.String("redeem-url", "", "Token redemption endpoint")
	flagSet.String("profile-url", "", "Profile access endpoint")
	flagSet.String("resource", "", "The resource that is protected (Azure AD only)")
	flagSet.String("validate-url", "", "Access token validation endpoint")
	flagSet.String("scope", "", "OAuth scope specification")
	flagSet.String("approval-prompt", "force", "OAuth approval_prompt")

	flagSet.String("signature-key", "", "GAP-Signature request signature key (algorithm:secretkey)")

	// These are options that allow you to tune various parameters for https://github.com/unrolled/secure
	flagSet.Var(&httpAllowedHosts, "httpAllowedHosts", "a list of fully qualified domain names that are allowed. Default is empty list, which allows any and all host names.")
	flagSet.Var(&httpHostsProxyHeaders, "httpHostsProxyHeaders", "a set of header keys that may hold a proxied hostname value for the request.")
	flagSet.Bool("httpSSLRedirect", false, "If set to true, then only allow HTTPS requests. Default is false.")
	flagSet.Bool("httpSSLTemporaryRedirect", false, "If true, then a 302 will be used while redirecting. Default is false (301).")
	flagSet.String("httpSSLHost", "", "the host name that is used to redirect HTTP requests to HTTPS. Default is \"\", which indicates to use the same host.")
	flagSet.Int64("httpSTSSeconds", 0, "The max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.")
	flagSet.Bool("httpSTSIncludeSubdomains", false, "If set to true, the 'includeSubdomains' will be appended to the Strict-Transport-Security header. Default is false.")
	flagSet.Bool("httpSTSPreload", false, "If set to true, the 'preload' flag will be appended to the Strict-Transport-Security header. Default is false.")
	flagSet.Bool("httpForceSTSHeader", false, "STS header is only included when the connection is HTTPS. If you want to force it to always be added, set to true. Default is false.")
	flagSet.Bool("httpFrameDeny", false, "If set to true, adds the X-Frame-Options header with the value of 'DENY'. Default is false.")
	flagSet.String("httpCustomFrameOptionsValue", "", "allows the X-Frame-Options header value to be set with a custom value. This overrides the FrameDeny option. Default is \"\".")
	flagSet.Bool("httpContentTypeNosniff", false, "If true, adds the X-Content-Type-Options header with the value 'nosniff'. Default is false.")
	flagSet.Bool("httpBrowserXssFilter", false, "If true, adds the X-XSS-Protection header with the value '1; mode=block'. Default is false.")
	flagSet.String("httpCustomBrowserXssValue", "", "Allows the X-XSS-Protection header value to be set with a custom value. This overrides the BrowserXssFilter option. Default is \"\".")
	flagSet.String("httpContentSecurityPolicy", "", "Allows the Content-Security-Policy header value to be set with a custom value. Default is \"\". Passing a template string will replace '$NONCE' with a dynamic nonce value of 16 bytes for each request which can be later retrieved using the Nonce function.")
	flagSet.String("httpPublicKey", "", "Implements HPKP to prevent MITM attacks with forged certificates. Default is \"\".")
	flagSet.String("httpReferrerPolicy", "", "Allows the Referrer-Policy header with the value to be set with a custom value. Default is \"\".")

	flagSet.Parse(os.Args[1:])

	if *showVersion {
		fmt.Printf("oauth2_proxy v%s (built with %s)\n", VERSION, runtime.Version())
		return
	}

	opts := NewOptions()

	cfg := make(EnvOptions)
	if *config != "" {
		_, err := toml.DecodeFile(*config, &cfg)
		if err != nil {
			log.Fatalf("ERROR: failed to load config file %s - %s", *config, err)
		}
	}
	cfg.LoadEnvForStruct(opts)
	options.Resolve(opts, flagSet, cfg)
	err := opts.Validate()
	if err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}

	validator := NewValidator(opts.EmailDomains, opts.AuthenticatedEmailsFile)
	oauthproxy := NewSecureProxy(opts, validator)

	s := &Server{
		Handler: LoggingHandler(os.Stdout, oauthproxy, opts.RequestLogging, opts.RequestLoggingFormat),
		Opts:    opts,
	}
	s.ListenAndServe()
}
