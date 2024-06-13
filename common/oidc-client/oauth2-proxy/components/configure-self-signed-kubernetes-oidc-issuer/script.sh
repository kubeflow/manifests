#!/bin/sh
set -e

KUBERNETES_API_SERVER_URL="${KUBERNETES_API_SERVER_URL:-https://kubernetes.default.svc}"
ISTIO_ROOT_NAMESPACE="${ISTIO_ROOT_NAMESPACE:-istio-system}"
REQUEST_AUTHENTICATION_NAME="${REQUEST_AUTHENTICATION_NAME:-m2m-token-issuer}"

RESOURCE_URL="\
${KUBERNETES_API_SERVER_URL}\
/apis/security.istio.io/v1/namespaces/\
${ISTIO_ROOT_NAMESPACE}\
/requestauthentications/\
${REQUEST_AUTHENTICATION_NAME}"

echo "Wait until resource RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} in namespace ${ISTIO_ROOT_NAMESPACE} is ready."
while true; do
  response="$(
    curl -s -o /dev/null \
      --url "${RESOURCE_URL}" \
      -w "%{http_code}" \
      --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
      --insecure
  )"
  if [ "${response}" = "200" ]; then
    break
  fi
  echo "Resource RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} in namespace ${ISTIO_ROOT_NAMESPACE} is not ready yet."
  sleep 5
done
echo "Resource RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} in namespace ${ISTIO_ROOT_NAMESPACE} is ready."

echo "Getting RequestAuthentication object."
REQUEST_AUTHENTICATION_OBJ="$(
  curl -s --request GET \
    --url "${RESOURCE_URL}" \
    --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
    --insecure
)"

# Get Issuer URL configured in RequestAuthentication.
ISSUER_URL="$(echo "${REQUEST_AUTHENTICATION_OBJ}" | awk -F'"' '/"issuer":/ { print $4 }')"
echo "Issuer Url in RequestAuthentication: ${ISSUER_URL}"

CURRENT_JWKS_ESCAPED="$(echo "${REQUEST_AUTHENTICATION_OBJ}" | awk -F'"' '/"jwks":/' | sed -n 's/^.*"jwks": "\(.*\)".*$/\1/p')"
printf "Current Jwks (escaped):\n%s\n" "${CURRENT_JWKS_ESCAPED}"

# GET URI to the JWKS.
JWKS_URI="$(
  curl -s --request GET \
    --url "${ISSUER_URL}/.well-known/openid-configuration" \
    --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
    --insecure |
    grep -o '"jwks_uri":"https:\/\/[^"]\+"' |
    sed 's/"jwks_uri":"\(.*\)"/\1/'
)"
echo "Jwks Uri from Well Known OpenID Configuration: ${JWKS_URI}"

# Get content of the JWKS.
JWKS="$(
  curl -s --request GET \
    --url "${JWKS_URI}" \
    --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
    --insecure
)"

# Format JWKS in a way that can be accepted in resource patch.
JWKS_ESCAPED="$(echo "${JWKS}" | sed 's/"/\\"/g')"
printf "JWKS from Well Known OpenID Configuration (escaped): \n%s\n" "${JWKS_ESCAPED}"

# If the JWKS from RequestAuthentication and OpenID Configuration is the same, don't to any changes.
if [ "$(echo "${JWKS_ESCAPED}" | base64 -w0)" = "$(echo "${CURRENT_JWKS_ESCAPED}" | base64 -w0)" ]; then
  echo "JWKS in RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} is configured correctly, exitting..."
  exit 0
fi
echo "JWKS in RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} needs to be configured."

# Patch the RequestAuthentication with JWKS.
curl -s --request PATCH \
  --url "${RESOURCE_URL}" \
  --header "Content-Type: application/json-patch+json" \
  --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
  -d '[{ "op": "add", "path": "/spec/jwtRules/0/jwks", "value": "'"${JWKS_ESCAPED}"'" }]' \
  --insecure
