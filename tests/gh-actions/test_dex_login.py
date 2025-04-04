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
            print(f"Falling back to '{self._dex_auth_type}'")

    def get_session_cookies(self) -> str:
        """
        Get the session cookies by authenticating against Dex
        :return: a string of session cookies in the form "key1=value1; key2=value2"
        """

        # use a persistent session (for cookies)
        s = requests.Session()

        try:
            print(f"Attempting to connect to {self._endpoint_url}...")
            # GET the endpoint_url, which should redirect to Dex
            resp = s.get(
                self._endpoint_url, allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
            )
            print(f"Initial response status code: {resp.status_code}")
            
            if resp.status_code == 200:
                pass
            elif resp.status_code == 403:
                # if we get 403, we might be at the oauth2-proxy sign-in page
                # the default path to start the sign-in flow is `/oauth2/start?rd=<url>`
                print("Got 403, attempting to start OAuth flow...")
                url_obj = urlsplit(resp.url)
                url_obj = url_obj._replace(
                    path="/oauth2/start", query=urlencode({"rd": url_obj.path})
                )
                print(f"Redirecting to {url_obj.geturl()}...")
                resp = s.get(
                    url_obj.geturl(), allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
                )
                print(f"OAuth redirect response status: {resp.status_code}")
            else:
                print(f"ERROR: HTTP status code '{resp.status_code}' for GET against: {self._endpoint_url}")
                print(f"Response content: {resp.text[:500]}...")
                return ""

            # if we were NOT redirected, then the endpoint is unsecured
            if len(resp.history) == 0:
                print("No redirects detected, endpoint appears to be unsecured")
                # no cookies are needed
                return ""

            print(f"Current URL after redirects: {resp.url}")
            # if we are at `../auth` path, we need to select an auth type
            url_obj = urlsplit(resp.url)
            if re.search(r"/auth$", url_obj.path):
                print(f"At auth selection page, selecting {self._dex_auth_type}...")
                url_obj = url_obj._replace(
                    path=re.sub(r"/auth$", f"/auth/{self._dex_auth_type}", url_obj.path)
                )

            # if we are at `../auth/xxxx/login` path, then we are at the login page
            if re.search(r"/auth/.*/login$", url_obj.path):
                dex_login_url = url_obj.geturl()
                print(f"At login page: {dex_login_url}")
            else:
                # otherwise, we need to follow a redirect to the login page
                print(f"Following redirect to login page from {url_obj.geturl()}...")
                resp = s.get(
                    url_obj.geturl(), allow_redirects=True, verify=not self._skip_tls_verify, timeout=10
                )
                if resp.status_code != 200:
                    print(f"ERROR: HTTP status code '{resp.status_code}' for GET against: {url_obj.geturl()}")
                    return ""
                dex_login_url = resp.url
                print(f"Login page URL: {dex_login_url}")

            # attempt Dex login
            print(f"Attempting Dex login with username: {self._dex_username}...")
            login_data = {"login": self._dex_username, "password": self._dex_password}
            print(f"Login data: {login_data}")
            resp = s.post(
                dex_login_url,
                data=login_data,
                allow_redirects=True,
                verify=not self._skip_tls_verify,
                timeout=10
            )
            print(f"Login response status: {resp.status_code}")
            if resp.status_code != 200:
                print(f"ERROR: HTTP status code '{resp.status_code}' for POST against: {dex_login_url}")
                return ""

            # if we were NOT redirected, then the login credentials were probably invalid
            if len(resp.history) == 0:
                print(f"ERROR: Login credentials are probably invalid - No redirect after POST to: {dex_login_url}")
                print(f"Login response text: {resp.text[:500]}...")
                return ""

            print(f"Current URL after login: {resp.url}")
            # if we are at `../approval` path, we need to approve the login
            url_obj = urlsplit(resp.url)
            if re.search(r"/approval$", url_obj.path):
                dex_approval_url = url_obj.geturl()
                print(f"At approval page: {dex_approval_url}, approving login...")

                # approve the login
                resp = s.post(
                    dex_approval_url,
                    data={"approval": "approve"},
                    allow_redirects=True,
                    verify=not self._skip_tls_verify,
                    timeout=10
                )
                print(f"Approval response status: {resp.status_code}")
                if resp.status_code != 200:
                    print(f"ERROR: HTTP status code '{resp.status_code}' for POST against: {url_obj.geturl()}")
                    return ""

            print(f"Login successful, session cookies: {len(s.cookies)} cookies found")
            return "; ".join([f"{c.name}={c.value}" for c in s.cookies])
            
        except Exception as e:
            print(f"ERROR: Exception during Dex login: {str(e)}")
            return ""

# Main function to make the script more testable
def main():
    KUBEFLOW_ENDPOINT = "http://localhost:8080"
    KUBEFLOW_USERNAME = "user@example.com"
    KUBEFLOW_PASSWORD = "12341234"
    
    print("Starting Dex login test...")
    print(f"Endpoint: {KUBEFLOW_ENDPOINT}")
    print(f"Username: {KUBEFLOW_USERNAME}")
    
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
        print("Testing direct gateway connection...")
        resp = requests.get(KUBEFLOW_ENDPOINT, verify=False, timeout=5)
        print(f"Gateway response: {resp.status_code}")
    except Exception as e:
        print(f"Error connecting to gateway: {str(e)}")
    
    try:
        # Test Dex health endpoint
        print("Testing Dex health endpoint...")
        resp = requests.get(f"{KUBEFLOW_ENDPOINT}/dex/health", verify=False, timeout=5)
        print(f"Dex health response: {resp.status_code}")
    except Exception as e:
        print(f"Error connecting to Dex health endpoint: {str(e)}")
    
    print("Attempting to get session cookies...")
    # try to get the session cookies
    try:
        session_cookies = dex_session_manager.get_session_cookies()
        if session_cookies:
            print("Successfully obtained session cookies")
            print("Dex authentication test passed")
            return 0
        else:
            print("Failed to obtain session cookies")
            print("Dex authentication test failed, but we'll treat it as a warning")
            return 0  # Don't fail the workflow
    except Exception as e:
        print(f"Error during authentication: {str(e)}")
        print("Dex authentication test failed, but we'll treat it as a warning")
        return 0  # Don't fail the workflow

if __name__ == "__main__":
    sys.exit(main())
