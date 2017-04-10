This is a simple, small container for the /bitly/oauth2_proxy.  All it does is uses the smallest possible container (scratch) and add in the certs for TLS.  Then it copies in the binary and runs it. That's it!  Smallest possible Docker container.

The Makefile can be versioned such that if you git checkout an old version and run make, then that specific version to be built.
