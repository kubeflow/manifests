#!/usr/bin/env python3

import re
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
        retry_delay = 2
        
        for attempt in range(max_retries):
            try:
                # use a persistent session (for cookies)
                s = requests.Session()
                
                # Add user-agent to avoid some security blocks or {} brackets also worked
                headers = {
                    'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36'
                }

                # GET the endpoint_url, which should redirect to Dex
                resp = s.get(
                    self._endpoint_url, 
                    headers=headers,
                    allow_redirects=True, 
                    verify=not self._skip_tls_verify,
                    timeout=30
                )
                
                if resp.status_code == 200:
                    pass
                elif resp.status_code == 403:
                    # Start OAuth2 flow explicitly
                    url_obj = urlsplit(resp.url)
                    oauth_url = f"{url_obj.scheme}://{url_obj.netloc}/oauth2/start?rd={url_obj.path or '/'}"
                    resp = s.get(
                        oauth_url,
                        headers=headers,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                        timeout=30
                    )
                else:
                    raise RuntimeError(
                        f"HTTP status code '{resp.status_code}' for GET against: {self._endpoint_url}"
                    )

                # if we were NOT redirected, then the endpoint is unsecured
                if len(resp.history) == 0:
                    # no cookies are needed
                    return ""

                # Navigate to auth path if needed
                url_obj = urlsplit(resp.url)
                if re.search(r"/auth$", url_obj.path):
                    auth_url = re.sub(r"/auth$", f"/auth/{self._dex_auth_type}", resp.url)
                    resp = s.get(
                        auth_url,
                        headers=headers,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                        timeout=30
                    )

                # Determine login URL
                if re.search(r"/auth/.*/login", urlsplit(resp.url).path):
                    dex_login_url = resp.url
                else:
                    # Get redirected to login page
                    resp = s.get(
                        resp.url,
                        headers=headers,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                        timeout=30
                    )
                    dex_login_url = resp.url

                # Check for and extract CSRF token
                login_data = {
                    "login": self._dex_username,
                    "password": self._dex_password
                }
                
                # Find and add CSRF token if present
                csrf_match = re.search(r'name="_csrf".*?value="([^"]+)"', resp.text)
                if csrf_match:
                    login_data["_csrf"] = csrf_match.group(1)
                    
                # Wait a moment before submitting the form (helps with some race conditions)
                import time
                time.sleep(1)
                    
                # attempt Dex login with proper headers
                resp = s.post(
                    dex_login_url,
                    data=login_data,
                    headers=headers,
                    allow_redirects=True,
                    verify=not self._skip_tls_verify,
                    timeout=30
                )
                
                # Handle 403 specifically - might need to restart oauth flow
                if resp.status_code == 403:
                    # Let's try one more approach - go through the oauth2 flow again
                    oauth_url = f"{urlsplit(self._endpoint_url).scheme}://{urlsplit(self._endpoint_url).netloc}/oauth2/start"
                    resp = s.get(
                        oauth_url,
                        headers=headers,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                        timeout=30
                    )
                    
                    # Continue with normal flow after restart
                    if resp.status_code == 200:
                        # If we have cookies now, we're good
                        if s.cookies:
                            return "; ".join([f"{c.name}={c.value}" for c in s.cookies])
                
                if resp.status_code != 200:
                    raise RuntimeError(
                        f"HTTP status code '{resp.status_code}' for POST against: {dex_login_url}"
                    )

                # if we are at `../approval` path, we need to approve the login
                url_obj = urlsplit(resp.url)
                if re.search(r"/approval$", url_obj.path):
                    dex_approval_url = url_obj.geturl()
                    # approve the login
                    resp = s.post(
                        dex_approval_url,
                        data={"approval": "approve"},
                        headers=headers,
                        allow_redirects=True,
                        verify=not self._skip_tls_verify,
                        timeout=30
                    )

                # Return the cookies if we have them
                if s.cookies:
                    return "; ".join([f"{c.name}={c.value}" for c in s.cookies])
                else:
                    raise RuntimeError("No cookies received after login flow")
                
            except Exception as e:
                if attempt < max_retries - 1:
                    print(f"Attempt {attempt+1} failed: {str(e)}")
                    print(f"Retrying in {retry_delay} seconds...")
                    import time
                    time.sleep(retry_delay)
                    retry_delay *= 2  # Exponential backoff
                else:
                    print(f"All {max_retries} attempts failed. Last error: {str(e)}")
                    raise

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
