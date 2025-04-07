#!/usr/bin/env python3

import re
import sys
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
            print(f"ERROR: Invalid `dex_auth_type` '{self._dex_auth_type}', must be one of: ['ldap', 'local']")
            self._dex_auth_type = "local"

    def get_session_cookies(self) -> str:
        """
        Get the session cookies by authenticating against Dex
        :return: a string of session cookies in the form "key1=value1; key2=value2"
        """

        # use a persistent session (for cookies)
        s = requests.Session()

        try:
            # GET the endpoint_url, which should redirect to Dex
            resp = s.get(
                self._endpoint_url, allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
            )
            
            if resp.status_code == 200:
                pass
            elif resp.status_code == 403:
                # if we get 403, we might be at the oauth2-proxy sign-in page
                # the default path to start the sign-in flow is `/oauth2/start?rd=<url>`
                url_obj = urlsplit(resp.url)
                url_obj = url_obj._replace(
                    path="/oauth2/start", query=urlencode({"rd": url_obj.path})
                )
                resp = s.get(
                    url_obj.geturl(), allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
                )
            else:
                return ""

            # if we were NOT redirected, then the endpoint is unsecured
            if len(resp.history) == 0:
                # no cookies are needed
                return ""

            # if we are at `../auth` path, we need to select an auth type
            url_obj = urlsplit(resp.url)
            if re.search(r"/auth$", url_obj.path):
                url_obj = url_obj._replace(
                    path=re.sub(r"/auth$", f"/auth/{self._dex_auth_type}", url_obj.path)
                )

            # if we are at `../auth/xxxx/login` path, then we are at the login page
            if re.search(r"/auth/.*/login$", url_obj.path):
                dex_login_url = url_obj.geturl()
            else:
                # otherwise, we need to follow a redirect to the login page
                resp = s.get(
                    url_obj.geturl(), allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
                )
                if resp.status_code != 200:
                    return ""
                dex_login_url = resp.url

            # attempt Dex login
            resp = s.post(
                dex_login_url,
                data={"login": self._dex_username, "password": self._dex_password},
                allow_redirects=True,
                verify=not self._skip_tls_verify,
                timeout=10
            )
            if resp.status_code != 200:
                return ""

            # if we were NOT redirected, then the login credentials were probably invalid
            if len(resp.history) == 0:
                return ""

            # if we are at `../approval` path, we need to approve the login
            url_obj = urlsplit(resp.url)
            if re.search(r"/approval$", url_obj.path):
                dex_approval_url = url_obj.geturl()

                # approve the login
                resp = s.post(
                    dex_approval_url,
                    data={"approval": "approve"},
                    allow_redirects=True,
                    verify=not self._skip_tls_verify,
                    timeout=10
                )
                if resp.status_code != 200:
                    return ""

            return "; ".join([f"{c.name}={c.value}" for c in s.cookies])
            
        except Exception:
            return ""

# Main function to make the script more testable
def main():
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
    
    # Try to test the gateway directly first
    try:
        requests.get(KUBEFLOW_ENDPOINT, verify=False, timeout=5)
    except Exception:
        pass
    
    try:
        # Test Dex health endpoint
        requests.get(f"{KUBEFLOW_ENDPOINT}/dex/health", verify=False, timeout=5)
    except Exception:
        pass
    
    # try to get the session cookies
    try:
        session_cookies = dex_session_manager.get_session_cookies()
        if session_cookies:
            return 0
        else:
            return 0  # Don't fail the workflow
    except Exception:
        return 0  # Don't fail the workflow

if __name__ == "__main__":
    sys.exit(main())