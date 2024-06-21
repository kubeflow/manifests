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

wait_for_resource_ready() {
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
    sleep 5
  done
}

get_request_authentication_obj() {
  curl -s --request GET \
    --url "${RESOURCE_URL}" \
    --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
    --insecure
}

get_issuer_url_from_obj() {
  obj="${1}"
  echo "${obj}" | awk -F'"' '/"issuer":/ { print $4 }'
}

get_current_escaped_jwks_from_obj() {
  obj="${1}"
  echo "${obj}" | awk -F'"' '/"jwks":/' | sed -n 's/^.*"jwks": "\(.*\)".*$/\1/p'
}

get_jwks_uri() {
  issuer_url="${1}"
  curl -s --request GET \
      --url "${issuer_url}/.well-known/openid-configuration" \
      --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
      --insecure |
      grep -o '"jwks_uri":"https:\/\/[^"]\+"' |
      sed 's/"jwks_uri":"\(.*\)"/\1/'
  }

get_jwks_from_uri() {
  jwks_uri="${1}"
  curl -s --request GET \
      --url "${jwks_uri}" \
      --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
      --insecure
}

# Format JWKS in a way that can be accepted in resource patch.
parse_escaped_jwks() {
  jwks="${1}"
  echo "${jwks}" | sed 's/"/\\"/g'
}

are_jwks_equal() {
  jwks1="${1}"
  jwks2="${2}"
  test "$(echo "${jwks1}" | base64 -w0)" = "$(echo "${jwks2}" | base64 -w0)"
}

patch_request_authentication_with_escaped_jwks() {
  jwks_escaped="${1}"
  curl -s --request PATCH \
    --url "${RESOURCE_URL}" \
    --header "Content-Type: application/json-patch+json" \
    --header "Authorization: Bearer $(cat /run/secrets/kubernetes.io/serviceaccount/token)" \
    -d '[{ "op": "add", "path": "/spec/jwtRules/0/jwks", "value": "'"${jwks_escaped}"'" }]' \
    --insecure
    echo
}

patch_request_authentication_with_jwks_if_required() {
  echo "Getting RequestAuthentication object."
  REQUEST_AUTHENTICATION_OBJ="$(get_request_authentication_obj)"

  ISSUER_URL="$(get_issuer_url_from_obj "${REQUEST_AUTHENTICATION_OBJ}")"
  echo "Issuer Url in RequestAuthentication: ${ISSUER_URL}"

  CURRENT_JWKS_ESCAPED="$(get_current_escaped_jwks_from_obj "${REQUEST_AUTHENTICATION_OBJ}")"
  printf "Current Jwks (escaped):\n%s\n" "${CURRENT_JWKS_ESCAPED}"

  JWKS_URI="$(get_jwks_uri "${ISSUER_URL}")"
  echo "Jwks Uri from Well Known OpenID Configuration: ${JWKS_URI}"

  JWKS="$(get_jwks_from_uri "${JWKS_URI}")"
  JWKS_ESCAPED="$(parse_escaped_jwks "${JWKS}")"
  printf "JWKS from Well Known OpenID Configuration (escaped): \n%s\n" "${JWKS_ESCAPED}"

  if are_jwks_equal "${JWKS_ESCAPED}" "${CURRENT_JWKS_ESCAPED}"; then
    echo "JWKS in RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} is configured correctly."
  else
    echo "JWKS in RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} needs to be configured."
    patch_request_authentication_with_escaped_jwks "${JWKS_ESCAPED}"
  fi
}

verify_jwks_in_request_authentication() {
  REQUEST_AUTHENTICATION_OBJ="$(get_request_authentication_obj)"
  ISSUER_URL="$(get_issuer_url_from_obj "${REQUEST_AUTHENTICATION_OBJ}")"
  CURRENT_JWKS_ESCAPED="$(get_current_escaped_jwks_from_obj "${REQUEST_AUTHENTICATION_OBJ}")"
  JWKS_URI="$(get_jwks_uri "${ISSUER_URL}")"
  JWKS="$(get_jwks_from_uri "${JWKS_URI}")"
  JWKS_ESCAPED="$(parse_escaped_jwks "${JWKS}")"
  if ! are_jwks_equal "${JWKS_ESCAPED}" "${CURRENT_JWKS_ESCAPED}"; then
    echo "JWKS not properly configured, exit with error code 1"
    exit 1
  fi
}

main() {
  echo "Wait until resource RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} in namespace ${ISTIO_ROOT_NAMESPACE} is ready."
  wait_for_resource_ready
  echo "Resource RequestAuthentication ${REQUEST_AUTHENTICATION_NAME} in namespace ${ISTIO_ROOT_NAMESPACE} is ready."

  echo "Patch RequestAuthentication with JWKS if required."
  patch_request_authentication_with_jwks_if_required

  echo "Wait 5 seconds before verifying RequestAuthentication JWKS configuration."
  sleep 5

  echo "Verify if RequestAuthentication is properly configured with JWKS..."
  verify_jwks_in_request_authentication
  echo "RequestAuthentication is properly configured with JWKS."
}

main
