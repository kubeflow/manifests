#!/usr/bin/env python3

import concurrent.futures
import re
import subprocess
import sys
import time
from dataclasses import dataclass
from urllib.parse import urlencode, urlsplit

import requests
import urllib3


ENDPOINT_URL = "http://localhost:8080"
DEX_USERNAME = "user@example.com"
DEX_PASSWORD = "12341234"
DEX_AUTH_TYPE = "local"
# Matches replicas: 2 in common/dex/base/deployment.yaml.
# Use a larger burst so the replica-distribution assertion is statistically stable in CI.
PARALLEL_SESSIONS = 8
# Dex authcode GC window: authcodes must be deleted after token exchange completes.
GC_WAIT_SECONDS = 90
REQUEST_TIMEOUT_SECONDS = 15
KUBECTL_TIMEOUT_SECONDS = 120
KUBECTL_REQUEST_TIMEOUT = "30s"

AUTHENTICATION_SUCCESS_LOG_MARKER = "login successful"
DEX_POD_SELECTOR = "app=dex"
DEX_AUTHCODE_RESOURCE = "authcodes.dex.coreos.com"


@dataclass
class ParallelAuthenticationResult:
    index: int
    ok: bool
    error: str = ""


class DexSessionManager:
    """
    This is a version of the KFPClientManager() which only generates the Dex session cookies.
    See https://www.kubeflow.org/docs/components/pipelines/user-guides/core-functions/connect-api/#kubeflow-platform---outside-the-cluster
    """

    def __init__(
        self,
        endpoint_url: str,
        dex_username: str,
        dex_password: str,
        dex_auth_type: str = "local",
        skip_tls_verify: bool = True,
    ):
        """
        Initialize the DexSessionManager

        :param endpoint_url: the Kubeflow Endpoint URL
        :param skip_tls_verify: if True, skip TLS verification
        :param dex_username: the Dex username
        :param dex_password: the Dex password
        :param dex_auth_type: the auth type to use if Dex has multiple enabled, one of: ['ldap', 'local']
        """
        self._endpoint_url = endpoint_url
        self._skip_tls_verify = skip_tls_verify
        self._dex_username = dex_username
        self._dex_password = dex_password
        self._dex_auth_type = dex_auth_type

        if self._skip_tls_verify:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

        # ensure `dex_auth_type` is valid
        if self._dex_auth_type not in ["ldap", "local"]:
            raise ValueError(
                f"Invalid `dex_auth_type` '{self._dex_auth_type}', must be one of: ['ldap', 'local']"
            )

    def _request_get(self, session: requests.Session, url: str) -> requests.Response:
        return session.get(
            url,
            allow_redirects=True,
            verify=not self._skip_tls_verify,
            timeout=REQUEST_TIMEOUT_SECONDS,
        )

    def _request_post(
        self, session: requests.Session, url: str, data: dict[str, str]
    ) -> requests.Response:
        return session.post(
            url,
            data=data,
            allow_redirects=True,
            verify=not self._skip_tls_verify,
            timeout=REQUEST_TIMEOUT_SECONDS,
        )

    @staticmethod
    def _has_oauth2_session_cookie(session: requests.Session) -> bool:
        return any(cookie.name.startswith("oauth2_proxy") for cookie in session.cookies)

    def _resolve_dex_login_url(self, session: requests.Session, url_object) -> str:
        """
        Given a URL object, navigate to the Dex login page and return its URL.
        Handles the optional /auth selector step before the /auth/<type>/login page.
        """
        # if we are at `../auth` path, we need to select an authentication type
        if re.search(r"/auth$", url_object.path):
            url_object = url_object._replace(
                path=re.sub(r"/auth$", f"/auth/{self._dex_auth_type}", url_object.path)
            )

        # if we are already at `../auth/xxxx/login`, use it directly
        if re.search(r"/auth/.*/login$", url_object.path):
            return url_object.geturl()

        # otherwise follow the redirect to the login page
        response = self._request_get(session, url_object.geturl())
        if response.status_code != 200:
            raise RuntimeError(
                f"HTTP status code '{response.status_code}' for GET against: {url_object.geturl()}"
            )
        return response.url

    def get_session_cookies(self) -> str:
        """
        Get the session cookies by authenticating against Dex.
        :return: a string of session cookies in the form "key1=value1; key2=value2"
        """
        session = requests.Session()

        try:
            # GET the endpoint URL, which should redirect to Dex
            response = self._request_get(session, self._endpoint_url)
            if response.status_code == 200:
                pass
            elif response.status_code in [401, 403]:
                # We may be at the oauth2-proxy sign-in page.
                # The standard path to start the sign-in flow is /oauth2/start?rd=<url>
                url_object = urlsplit(response.url)
                url_object = url_object._replace(
                    path="/oauth2/start",
                    query=urlencode({"rd": url_object.path}),
                )
                response = self._request_get(session, url_object.geturl())
                if response.status_code not in [200, 302]:
                    raise RuntimeError(
                        f"HTTP status code '{response.status_code}' for GET against oauth2/start"
                    )
            else:
                raise RuntimeError(
                    f"HTTP status code '{response.status_code}' for GET against: {self._endpoint_url}"
                )

            # if we were NOT redirected, the endpoint is unsecured — no cookies needed
            if len(response.history) == 0:
                return ""

            dex_login_url = self._resolve_dex_login_url(session, urlsplit(response.url))

            # submit the login credentials
            response = self._request_post(
                session,
                dex_login_url,
                data={"login": self._dex_username, "password": self._dex_password},
            )

            if response.status_code == 403:
                # 403 after login POST can mean the oauth2-proxy session expired mid-flow.
                # If the redirect chain passed through /oauth2/callback and we already have
                # a valid oauth2 session cookie, we are actually authenticated — return early.
                history_urls = [h.url for h in response.history]
                if (
                    any("/oauth2/callback" in u for u in history_urls)
                    and self._has_oauth2_session_cookie(session)
                ):
                    return "; ".join(
                        [f"{cookie.name}={cookie.value}" for cookie in session.cookies]
                    )

                # Otherwise restart the oauth2 flow and retry the login once
                oauth_url = (
                    f"{urlsplit(self._endpoint_url).scheme}://"
                    f"{urlsplit(self._endpoint_url).netloc}/oauth2/start"
                )
                response = self._request_get(session, oauth_url)
                if response.status_code not in [200, 302]:
                    raise RuntimeError(
                        "HTTP status code "
                        f"'{response.status_code}' for GET against oauth2/start during 403 recovery"
                    )

                dex_login_url = self._resolve_dex_login_url(session, urlsplit(response.url))
                response = self._request_post(
                    session,
                    dex_login_url,
                    data={"login": self._dex_username, "password": self._dex_password},
                )

            if response.status_code != 200:
                raise RuntimeError(
                    f"HTTP status code '{response.status_code}' for POST against: {dex_login_url}"
                )

            # no redirect after login POST means credentials were invalid
            if len(response.history) == 0:
                raise RuntimeError(
                    "Authentication credentials are probably invalid - "
                    f"no redirect after POST to: {dex_login_url}"
                )

            # if we are at `../approval` path, we need to approve the login
            url_object = urlsplit(response.url)
            if re.search(r"/approval$", url_object.path):
                dex_approval_url = url_object.geturl()
                response = self._request_post(
                    session, dex_approval_url, data={"approval": "approve"}
                )
                if response.status_code != 200:
                    raise RuntimeError(
                        f"HTTP status code '{response.status_code}' for POST against: {url_object.geturl()}"
                    )

            return "; ".join([f"{cookie.name}={cookie.value}" for cookie in session.cookies])

        except requests.RequestException as exc:
            raise RuntimeError(f"Dex authentication request failed: {exc}") from exc


def run_cmd(cmd: list[str], timeout: int = KUBECTL_TIMEOUT_SECONDS) -> subprocess.CompletedProcess:
    try:
        return subprocess.run(
            cmd,
            check=False,
            text=True,
            capture_output=True,
            timeout=timeout,
        )
    except subprocess.TimeoutExpired as exc:
        raise RuntimeError(f"Command timed out after {timeout}s: {' '.join(cmd)}") from exc


def run_cmd_or_fail(cmd: list[str], timeout: int = KUBECTL_TIMEOUT_SECONDS) -> subprocess.CompletedProcess:
    result = run_cmd(cmd, timeout=timeout)
    if result.returncode != 0:
        raise RuntimeError(
            "Command failed "
            f"(rc={result.returncode}): {' '.join(cmd)}\n"
            f"stdout:\n{result.stdout.strip()}\n"
            f"stderr:\n{result.stderr.strip()}"
        )
    return result


def get_dex_pods(min_replicas: int = 2) -> list[str]:
    """
    Return the names of running Dex pods in the auth namespace.
    Raises if fewer than min_replicas pods are found — the parallel authentication
    test requires at least two replicas to verify cross-replica load distribution.
    """
    cmd = [
        "kubectl",
        "--request-timeout", KUBECTL_REQUEST_TIMEOUT,
        "-n", "auth",
        "get", "pods",
        "-l", DEX_POD_SELECTOR,
        "-o", "jsonpath={.items[*].metadata.name}",
    ]
    result = run_cmd_or_fail(cmd)
    pods = [pod for pod in result.stdout.strip().split() if pod]
    if len(pods) < min_replicas:
        raise RuntimeError(
            f"Expected at least {min_replicas} Dex pods (selector: {DEX_POD_SELECTOR}) "
            f"in namespace auth, found: {pods}. "
            "The Dex deployment at common/dex/base/deployment.yaml is configured with "
            "replicas: 2 — ensure all pods have reached the Ready state before running this test."
        )
    return pods


def count_authentication_hits_for_pod(pod: str, since_seconds: int) -> int:
    """Count how many successful authentication events appear in a pod's logs."""
    cmd = [
        "kubectl",
        "--request-timeout", KUBECTL_REQUEST_TIMEOUT,
        "-n", "auth",
        "logs", pod,
        f"--since={since_seconds}s",
    ]
    result = run_cmd_or_fail(cmd)
    return len(re.findall(re.escape(AUTHENTICATION_SUCCESS_LOG_MARKER), result.stdout))


def count_authcodes_objects() -> int:
    """
    Count the number of Dex authcode CRD objects currently in the cluster.
    Dex creates one authcode object per login; the GC process deletes them after
    the token exchange completes. Returns 0 if no instances exist.
    """
    cmd = [
        "kubectl",
        "--request-timeout", KUBECTL_REQUEST_TIMEOUT,
        "get", DEX_AUTHCODE_RESOURCE,
        "-A", "--no-headers",
    ]
    result = run_cmd(cmd)
    # "no resources found" is a normal state — return 0 rather than raising
    if result.returncode != 0:
        combined = (result.stdout + "\n" + result.stderr).lower()
        if "no resources found" in combined:
            return 0
        raise RuntimeError(
            f"Failed to query {DEX_AUTHCODE_RESOURCE}: {result.stderr.strip()}"
        )
    return len([line for line in result.stdout.splitlines() if line.strip()])


def run_single_authentication() -> str:
    manager = DexSessionManager(
        endpoint_url=ENDPOINT_URL,
        skip_tls_verify=True,
        dex_username=DEX_USERNAME,
        dex_password=DEX_PASSWORD,
        dex_auth_type=DEX_AUTH_TYPE,
    )
    return manager.get_session_cookies()


def run_parallel_authentication_session(index: int) -> ParallelAuthenticationResult:
    try:
        run_single_authentication()
        return ParallelAuthenticationResult(index=index, ok=True)
    except Exception as exc:
        return ParallelAuthenticationResult(index=index, ok=False, error=str(exc))


def run_parallel_validation() -> None:
    """
    Validates that:
    1. PARALLEL_SESSIONS concurrent authentication sessions all succeed against a
       multi-replica Dex deployment.
    2. Authentication traffic is distributed across at least two Dex replicas (load balancer
       is working). With no sessionAffinity on the Dex Service, the Kubernetes load
       balancer distributes connections freely, so a single burst is sufficient to
       observe both replicas receiving traffic.
    3. Dex authcode CRD objects created during the burst are garbage collected after
       the GC_WAIT_SECONDS window. With storage.type=kubernetes, authcodes are
       Kubernetes CRD objects that Dex actively deletes after each token exchange.

    Requires at least 2 Dex replicas (replicas: 2 in common/dex/base/deployment.yaml).
    The since_seconds log window is sized to cover the burst plus GC wait plus a buffer
    so that baseline and post-burst reads always observe the same window.
    """
    pods = get_dex_pods(min_replicas=2)
    print(f"Dex pods: {pods}")

    # Size the log window to cover the burst duration plus GC wait plus a buffer.
    since_seconds = max(GC_WAIT_SECONDS + 120, 300)

    # Snapshot state before the burst
    baseline_hits = {
        pod: count_authentication_hits_for_pod(pod, since_seconds)
        for pod in pods
    }
    authcodes_before = count_authcodes_objects()

    print(f"Running parallel authentication burst with sessions={PARALLEL_SESSIONS}")

    # Run all parallel authentication sessions and collect results
    failures = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=PARALLEL_SESSIONS) as executor:
        futures = [
            executor.submit(run_parallel_authentication_session, index)
            for index in range(PARALLEL_SESSIONS)
        ]
        for future in concurrent.futures.as_completed(futures, timeout=REQUEST_TIMEOUT_SECONDS * 3):
            result = future.result()
            if not result.ok:
                failures.append(result)

    if failures:
        error_summary = "; ".join(
            [f"session={f.index} error={f.error}" for f in failures]
        )
        raise RuntimeError(
            f"Parallel authentication session failures: {error_summary}"
        )

    # Verify that at least two distinct replicas handled authentication requests.
    # This confirms the load balancer is distributing traffic across pods.
    # Requires sessionAffinity to be absent from the Dex Service — affinity would pin
    # all sessions from the same source IP to a single pod, defeating this check.
    post_hits = {
        pod: count_authentication_hits_for_pod(pod, since_seconds)
        for pod in pods
    }
    hit_delta = {
        pod: max(post_hits[pod] - baseline_hits[pod], 0)
        for pod in pods
    }
    print(f"Authentication hit delta by pod: {hit_delta}")

    hit_pods = [pod for pod, delta in hit_delta.items() if delta > 0]
    if len(hit_pods) < 2:
        raise RuntimeError(
            "Expected authentication traffic across at least two Dex replicas "
            f"but observed: {hit_delta}. "
            "Verify that the Dex Service has no sessionAffinity configured."
        )

    # Verify GC: authcodes created during the burst must be cleaned up after the wait window.
    # Dex creates one authcode CRD object per login and deletes it after the token exchange.
    # If GC is broken, authcodes accumulate indefinitely.
    authcodes_after_burst = count_authcodes_objects()
    print(f"Authcodes count: before={authcodes_before} after_burst={authcodes_after_burst}")

    time.sleep(GC_WAIT_SECONDS)
    authcodes_after_wait = count_authcodes_objects()
    print(f"Authcodes count after GC wait ({GC_WAIT_SECONDS}s): {authcodes_after_wait}")

    if authcodes_after_burst > authcodes_before:
        # The burst created new authcodes — GC must reduce the count
        if authcodes_after_wait >= authcodes_after_burst:
            raise RuntimeError(
                "Authcodes did not decrease after GC wait window — "
                "Dex GC may not be functioning correctly. "
                f"before={authcodes_before} burst={authcodes_after_burst} "
                f"after_wait={authcodes_after_wait}"
            )
    elif authcodes_after_wait > authcodes_after_burst:
        # No burst growth but count increased during wait — unexpected leak
        raise RuntimeError(
            "Authcodes increased during GC wait window despite no observed burst growth — "
            "possible authcode leak from another process. "
            f"before={authcodes_before} burst={authcodes_after_burst} "
            f"after_wait={authcodes_after_wait}"
        )


def main() -> None:
    run_single_authentication()
    print("Dex single authentication validation passed")

    run_parallel_validation()
    print("Dex parallel authentication and GC validation passed")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"Dex authentication test failed: {exc}", file=sys.stderr)
        raise SystemExit(1)
