#!/usr/bin/env python3

import re
import time
from urllib.parse import urlsplit, urlencode
import requests
import urllib3


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
        skip_tls_verify: bool = False,
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
        self._client = None

        # disable SSL verification, if requested
        if self._skip_tls_verify:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

        # ensure `dex_default_auth_type` is valid
        if self._dex_auth_type not in ["ldap", "local"]:
            raise ValueError(
                f"Invalid `dex_auth_type` '{self._dex_auth_type}', must be one of: ['ldap', 'local']"
            )

    def get_session_cookies(self) -> str:
        """
        Get the session cookies by authenticating against Dex
        :return: a string of session cookies in the form "key1=value1; key2=value2"
        """
        max_retries = 3
        base_retry_delay = 2

        for attempt in range(max_retries):
            # Create a fresh session for each attempt to avoid stale state
            session = requests.Session()
            
            if attempt > 0:
                delay = base_retry_delay * (2 ** (attempt - 1))
                time.sleep(delay)
            
            try:
                # GET the endpoint_url, which should redirect to Dex
                response = session.get(
                    self._endpoint_url,
                    allow_redirects=True,
                    verify=not self._skip_tls_verify
                )
                if response.status_code == 200:
                    pass
                elif response.status_code in [401, 403]:
                    # if we get 401/403, we might be at the oauth2-proxy sign-in page
                    # the default path to start the sign-in flow is `/oauth2/start?rd=<url>`
                    url_object = urlsplit(response.url)
                    url_object = url_object._replace(
                        path="/oauth2/start",
                        query=urlencode({"rd": url_object.path})
                    )
                    response = session.get(
                        url_object.geturl(),
                        allow_redirects=True,
                        verify=not self._skip_tls_verify
                    )
                    if response.status_code not in [200, 302]:
                        raise RuntimeError(
                            f"HTTP status code '{response.status_code}' for GET against oauth2/start"
                        )
                else:
                    raise RuntimeError(
                        f"HTTP status code '{response.status_code}' for GET against: {self._endpoint_url}"
                    )

                # if we were NOT redirected, then the endpoint is unsecured
                if len(response.history) == 0:
                    # No cookies are needed
                    return ""

                # if we are at `../auth` path, we need to select an auth type
                url_object = urlsplit(response.url)
                if re.search(r"/auth$", url_object.path):
                    url_object = url_object._replace(
                        path=re.sub(r"/auth$", f"/auth/{self._dex_auth_type}", url_object.path)
                    )

                # if we are at `../auth/xxxx/login` path, then we are at the login page
                if re.search(r"/auth/.*/login$", url_object.path):
                    dex_login_url = url_object.geturl()
                else:
                    # otherwise, we need to follow a redirect to the login page
                    response = session.get(
                        url_object.geturl(),
                        allow_redirects=True,
                        verify=not self._skip_tls_verify
                    )
                    if response.status_code != 200:
                        raise RuntimeError(
                            f"HTTP status code '{response.status_code}' for GET against: {url_object.geturl()}"
                        )
                    dex_login_url = response.url

                # attempt Dex login
                response = session.post(
                    dex_login_url,
                    data={"login": self._dex_username, "password": self._dex_password},
                    allow_redirects=True,
                    verify=not self._skip_tls_verify,
                )

                # Handle 403 specifically - might need to restart oauth flow
                if response.status_code == 403:
                    # Try one more approach - go through the oauth2 flow again
                    oauth_url = f"{urlsplit(self._endpoint_url).scheme}://{urlsplit(self._endpoint_url).netloc}/oauth2/start"
                    response = session.get(
                        oauth_url,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                    )
                    # Continue with normal flow after restart
                    if response.status_code == 200 and session.cookies:
                        return "; ".join([f"{c.name}={c.value}" for c in session.cookies])

                if response.status_code != 200:
                    raise RuntimeError(
                        f"HTTP status code '{response.status_code}' for POST against: {dex_login_url}"
                    )

                # if we were NOT redirected, then the login credentials were probably invalid
                if len(response.history) == 0:
                    raise RuntimeError(
                        f"Login credentials are probably invalid - "
                        f"No redirect after POST to: {dex_login_url}"
                    )

                # if we are at `../approval` path, we need to approve the login
                url_object = urlsplit(response.url)
                if re.search(r"/approval$", url_object.path):
                    dex_approval_url = url_object.geturl()
                    # Approve the login
                    response = session.post(
                        dex_approval_url,
                        data={"approval": "approve"},
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                    )
                    if response.status_code != 200:
                        raise RuntimeError(
                            f"HTTP status code '{response.status_code}' for POST against: {url_object.geturl()}"
                        )

                return "; ".join([f"{c.name}={c.value}" for c in session.cookies])

            except Exception as e:
                if attempt == max_retries - 1:  # Last attempt
                    print(f"All {max_retries} attempts failed. Last error: {str(e)}")
                    raise
                next_delay = base_retry_delay * (2 ** attempt)
                print(f"Attempt {attempt + 1} failed: {str(e)}")


KUBEFLOW_ENDPOINT = "http://localhost:8080"
KUBEFLOW_USERNAME = "user@example.com"
KUBEFLOW_PASSWORD = "12341234"

# initialize a DexSessionManager
dex_session_manager = DexSessionManager(
    endpoint_url=KUBEFLOW_ENDPOINT,
    skip_tls_verify=True,
    dex_username=KUBEFLOW_USERNAME,
    dex_password=KUBEFLOW_PASSWORD,
    dex_auth_type="local",
)

# try to get the session cookies
# NOTE: this will raise an exception if something goes wrong
session_cookies = dex_session_manager.get_session_cookies()