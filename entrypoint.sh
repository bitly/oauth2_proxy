#!/bin/bash
set -e

export PATH=/bin:/usr/bin:/usr/local/bin:/usr/sbin:/

if [ ! -z ${OAUTH2_PROXY_CLIENT_ID+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -client-id ${OAUTH2_PROXY_CLIENT_ID}"
fi

if [ ! -z ${OAUTH2_PROXY_CLIENT_SECRET+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -client-secret ${OAUTH2_PROXY_CLIENT_SECRET}"
fi

if [ ! -z ${OAUTH2_PROXY_COOKIE_EXPIRE+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -cookie-expire ${OAUTH2_PROXY_COOKIE_EXPIRE}"
fi

if [ ! -z ${OAUTH2_PROXY_COOKIE_SECRET+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -cookie-secret ${OAUTH2_PROXY_COOKIE_SECRET}"
fi

if [ ! -z ${OAUTH2_PROXY_EMAIL_DOMAIN+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -email-domain \"${OAUTH2_PROXY_EMAIL_DOMAIN}\""
fi

if [ ! -z ${OAUTH2_PROXY_GITHUB_TEAM+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -github-team ${OAUTH2_PROXY_GITHUB_TEAM}"
fi

if [ ! -z ${OAUTH2_PROXY_GITHUB_ORG+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -github-org ${OAUTH2_PROXY_GITHUB_ORG}"
fi

if [ ! -z ${OAUTH2_PROXY_HTTP_ADDRESS+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -http-address ${OAUTH2_PROXY_HTTP_ADDRESS}"
fi

if [ ! -z ${OAUTH2_PROXY_HTTPS_ADDRESS+x} ] && [ ! -z $OAUTH2_PROXY_HTTPS_ADDRESS ]; then
   PROXY_ARGS="${PROXY_ARGS} -https-address ${OAUTH2_PROXY_HTTPS_ADDRESS}"
fi

if [ ! -z ${OAUTH2_PROXY_REDIRECT_URL+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -redirect-url ${OAUTH2_PROXY_REDIRECT_URL}"
fi

if [ ! -z ${OAUTH2_PROXY_TLS_CERT+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -tls-cert ${OAUTH2_PROXY_TLS_CERT}"
fi

if [ ! -z ${OAUTH2_PROXY_TLS_KEY+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -tls-key ${OAUTH2_PROXY_TLS_KEY}"
fi

if [ ! -z ${OAUTH2_PROXY_PROVIDER+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -provider ${OAUTH2_PROXY_PROVIDER}"
fi

if [ ! -z ${OAUTH2_PROXY_UPSTREAM+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -upstream ${OAUTH2_PROXY_UPSTREAM}"
fi

if [ ! -z ${OAUTH2_PROXY_SIGN_AWS_REQUEST_REGION+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -sign-aws-request-region ${OAUTH2_PROXY_SIGN_AWS_REQUEST_REGION}"
fi

if [ ! -z ${OAUTH2_PROXY_SIGN_AWS_REQUEST_SERVICE+x} ]; then
   PROXY_ARGS="${PROXY_ARGS} -sign-aws-request-service ${OAUTH2_PROXY_SIGN_AWS_REQUEST_SERVICE}"
fi

if [ ! -z ${OAUTH2_PROXY_AWS_ACCESS_KEY+x} ]; then
   AWS_ACCESS_KEY=$OAUTH2_PROXY_AWS_ACCESS_KEY
fi

if [ ! -z ${OAUTH2_PROXY_AWS_SECRET_ACCESS_KEY+x} ]; then
   AWS_SECRET_ACCESS_KEY=$OAUTH2_PROXY_AWS_SECRET_ACCESS_KEY
fi 

echo "Launching oauth2_proxy..."
exec /gosu nobody /oauth2_proxy ${PROXY_ARGS}
