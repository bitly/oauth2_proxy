# for heroku support (along with .godir)
web: google_auth_proxy --http-address=0.0.0.0:$PORT --upstream=$UPSTREAM --redirect-url=$REDIRECT_URL --google-apps-domain=$GOOGLE_APPS_DOMAIN --cookie-domain=$COOKIE_DOMAIN --cookie-secret=$COOKIE_SECRET --client-secret=$CLIENT_SECRET --client-id=$CLIENT_ID ${PASS_SECRET:+--pass-secret=$PASS_SECRET} ${REWRITE_HOST:+--rewrite-host=$REWRITE_HOST}
