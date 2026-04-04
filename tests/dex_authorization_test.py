#!/usr/bin/env python3
"""
Authorization checks for the Dex + oauth2-proxy stack.

This test reuses the Dex login flow from dex_login_test.py, then verifies that
the volumes web app honors namespace-scoped authorization decisions.
"""

import sys

import requests
import urllib3

from dex_login_test import DexSessionManager, run_cmd_or_fail

ENDPOINT_URL = "http://localhost:8080"
DEX_USERNAME = "user@example.com"
DEX_PASSWORD = "12341234"
DEX_AUTH_TYPE = "local"

AUTHORIZED_NAMESPACE = "kubeflow-user-example-com"
UNAUTHORIZED_NAMESPACE = "default"

VOLUMES_UI_PATH = "/volumes/"
VOLUMES_API_TEMPLATE = "/volumes/api/namespaces/{namespace}/pvcs"

REQUEST_TIMEOUT_SECONDS = 15

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


def _request(
    url: str,
    headers: dict[str, str],
    label: str,
    expected_codes: list[int],
    allow_redirects: bool = True,
) -> requests.Response:
    response = requests.get(
        url,
        headers=headers,
        allow_redirects=allow_redirects,
        verify=False,
        timeout=REQUEST_TIMEOUT_SECONDS,
    )
    print(f"{label}: HTTP {response.status_code}")
    if response.status_code not in expected_codes:
        raise RuntimeError(
            f"{label}: expected one of {expected_codes}, got {response.status_code}. "
            f"Response body (first 300 chars): {response.text[:300]}"
        )
    return response


def _session_cookie_header(session_cookies: str, xsrf_token: str | None = None) -> dict[str, str]:
    cookie_header = session_cookies
    headers = {"Cookie": cookie_header}
    if xsrf_token:
        headers["Cookie"] = f"{cookie_header}; XSRF-TOKEN={xsrf_token}"
        headers["X-XSRF-TOKEN"] = xsrf_token
    return headers


def _get_xsrf_token(session_cookies: str) -> str:
    response = _request(
        url=f"{ENDPOINT_URL}{VOLUMES_UI_PATH}",
        headers=_session_cookie_header(session_cookies),
        label="Get volumes UI XSRF token",
        expected_codes=[200],
    )
    xsrf_token = response.cookies.get("XSRF-TOKEN")
    if not xsrf_token:
        raise RuntimeError("Volumes UI did not return an XSRF-TOKEN cookie")
    return xsrf_token


def _volumes_api_url(namespace: str) -> str:
    return f"{ENDPOINT_URL}{VOLUMES_API_TEMPLATE.format(namespace=namespace)}"


def run_authorized_access_validation(session_cookies: str, xsrf_token: str) -> None:
    _request(
        url=_volumes_api_url(AUTHORIZED_NAMESPACE),
        headers=_session_cookie_header(session_cookies, xsrf_token),
        label="Probe 1 (authorized cookie, own namespace)",
        expected_codes=[200],
    )
    print("PASS: Authorized Dex session can access the volumes API in its own namespace")


def run_unauthenticated_access_validation() -> None:
    response = _request(
        url=_volumes_api_url(AUTHORIZED_NAMESPACE),
        headers={},
        label="Probe 2 (no cookie)",
        expected_codes=[302, 401, 403],
        allow_redirects=False,
    )
    if response.status_code == 200:
        raise RuntimeError("Unauthenticated request unexpectedly reached a protected endpoint")
    print("PASS: Unauthenticated request is rejected before namespace access is granted")


def run_cross_namespace_authorization_validation(
    session_cookies: str, xsrf_token: str
) -> None:
    _request(
        url=_volumes_api_url(UNAUTHORIZED_NAMESPACE),
        headers=_session_cookie_header(session_cookies, xsrf_token),
        label=f"Probe 3 (authorized cookie, unauthorized namespace={UNAUTHORIZED_NAMESPACE})",
        expected_codes=[403],
    )
    print("PASS: Authenticated user is denied access to a different namespace")


def run_unauthorized_serviceaccount_validation() -> None:
    unauthorized_token = run_cmd_or_fail(
        ["kubectl", "-n", UNAUTHORIZED_NAMESPACE, "create", "token", "default"]
    ).stdout.strip()
    _request(
        url=_volumes_api_url(AUTHORIZED_NAMESPACE),
        headers={"Authorization": f"Bearer {unauthorized_token}"},
        label=(
            f"Probe 4 (unauthorized ServiceAccount token from "
            f"namespace={UNAUTHORIZED_NAMESPACE})"
        ),
        expected_codes=[401, 403],
        allow_redirects=False,
    )
    print("PASS: ServiceAccount token from an unauthorized namespace is rejected")


def main() -> None:
    print("Obtaining Dex session cookie for authorization probes...")
    manager = DexSessionManager(
        endpoint_url=ENDPOINT_URL,
        skip_tls_verify=True,
        dex_username=DEX_USERNAME,
        dex_password=DEX_PASSWORD,
        dex_auth_type=DEX_AUTH_TYPE,
    )
    session_cookies = manager.get_session_cookies()
    xsrf_token = _get_xsrf_token(session_cookies)

    run_authorized_access_validation(session_cookies, xsrf_token)
    run_unauthenticated_access_validation()
    run_cross_namespace_authorization_validation(session_cookies, xsrf_token)
    run_unauthorized_serviceaccount_validation()

    print("\nAll authorization probes passed")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"Dex authorization test failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
