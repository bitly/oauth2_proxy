#!/bin/sh
set -e

ISSUER="https://preprod.idpdecathlon.oxylane.com"

####################################################################
#                         UTILS FUNCTIONS                          #
####################################################################
missingParameter() {
    printf "Required parameter %s not set!\n" $1
    exit 1
}

####################################################################
#                   CHECKING INPUT PARAMETERS                      #
####################################################################

if [ -z $OAUTH2_PROXY_CLIENT_ID ]; then
  missingParameter "OAUTH2_PROXY_CLIENT_ID"
fi

if [ -z $OAUTH2_PROXY_CLIENT_SECRET ]; then
  missingParameter "OAUTH2_PROXY_CLIENT_SECRET"
fi

if [ -z $OAUTH2_PROXY_COOKIE_SECRET ]; then
  missingParameter "OAUTH2_PROXY_COOKIE_SECRET"
fi

if [ "${DECATHLON_ENV}" == "PRODUCTION" ]; then
    ISSUER="https://idpdecathlon.oxylane.com"
else
    ISSUER="${ISSUER} --ssl-insecure-skip-verify=true --cookie-secure=false"
fi


####################################################################
#                        MANAGING COMMAND                          #
####################################################################
CMD="/usr/local/bin/oauth2_proxy"
CMD="$CMD --provider=decathlon"
CMD="$CMD --oidc-issuer-url=${ISSUER}"
CMD="$CMD --email-domain=\"decathlon.com\""
CMD="$CMD --email-domain=\"oxylane.com\""
CMD="$CMD --skip-provider-button=true"
CMD="$CMD --cookie-secret=\"${OAUTH2_PROXY_COOKIE_SECRET}\""
CMD="$CMD --client-id=${OAUTH2_PROXY_CLIENT_ID}"
CMD="$CMD --client-secret=${OAUTH2_PROXY_CLIENT_SECRET}"
CMD="$CMD --http-address=0.0.0.0:4180"
CMD="$CMD $1"

echo "$CMD"
eval $CMD
