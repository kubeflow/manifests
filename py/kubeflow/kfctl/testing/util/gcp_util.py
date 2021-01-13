import datetime
import logging
import os
from time import sleep
from google.auth.transport.requests import Request
from google.oauth2 import id_token

import requests
from retrying import retry
from requests.exceptions import SSLError
from requests.exceptions import ConnectionError as ReqConnectionError

COOKIE_NAME = "KUBEFLOW-AUTH-KEY"

def may_get_env_var(name):
  env_val = os.getenv(name)
  if env_val:
    logging.info("%s is set", name)
    return env_val

  raise Exception("%s not set" % name)

# Code copied from:
# https://cloud.google.com/iap/docs/authentication-howto#iap_make_request-python
def make_iap_request(url, client_id, method='GET', **kwargs):
    """Makes a request to an application protected by Identity-Aware Proxy.

    Args:
      url: The Identity-Aware Proxy-protected URL to fetch.
      client_id: The client ID used by Identity-Aware Proxy.
      method: The request method to use
              ('GET', 'OPTIONS', 'HEAD', 'POST', 'PUT', 'PATCH', 'DELETE')
      **kwargs: Any of the parameters defined for the request function:
                https://github.com/requests/requests/blob/master/requests/api.py
                If no timeout is provided, it is set to 90 by default.

    Returns:
      The page body, or raises an exception if the page couldn't be retrieved.
    """
    # Set the default timeout, if missing
    if 'timeout' not in kwargs:
        kwargs['timeout'] = 90

    # Obtain an OpenID Connect (OIDC) token from metadata server or using service
    # account.
    google_open_id_connect_token = id_token.fetch_id_token(Request(), client_id)

    # Fetch the Identity-Aware Proxy-protected URL, including an
    # Authorization header containing "Bearer " followed by a
    # Google-issued OpenID Connect token for the service account.
    resp = requests.request(
        method, url,
        headers={'Authorization': 'Bearer {}'.format(
            google_open_id_connect_token)}, **kwargs)
    if resp.status_code == 403: # pylint: disable=no-else-raise
        raise Exception('Service account does not have permission to '
                        'access the IAP-protected application.')
    elif resp.status_code != 200: # pylint: disable=no-else-raise
        raise Exception(
            'Bad response from application: {!r} / {!r} / {!r}'.format(
                resp.status_code, resp.headers, resp.text))
    else:
        return resp.text

def iap_is_ready(url, wait_min=15):
  """
  Checks if the kubeflow endpoint is ready.

  Args:
    url: The url endpoint
  Returns:
    True if the url is ready
  """
  client_id = may_get_env_var("CLIENT_ID")
  # Wait up to 30 minutes for IAP access test.
  num_req = 0
  end_time = datetime.datetime.now() + datetime.timedelta(
      minutes=wait_min)
  while datetime.datetime.now() < end_time:
    num_req += 1
    logging.info("Trying url: %s", url)
    try:
      resp = make_iap_request(url, client_id, method="GET", verify=False)
      logging.info("Response: %s", resp)
      logging.info("Endpoint is ready for %s!", url)
      return True
    except Exception as e: # pylint: disable=broad-except
      logging.info("%s: Endpoint not ready, exception caught %s, request "
                   "number: %s", url, str(e), num_req)
    sleep(10)
  return False

def _send_req(wait_sec, url, req_gen, retry_result_code=None):
  """ Helper function to send requests and retry when the endpoint is not ready.

  Args:
  wait_sec: int, max time to wait and retry in seconds.
  url: str, url to send the request, used only for logging.
  req_gen: lambda, no parameter function to generate requests.Request for the
  function to send to the endpoint.
  retry_result_code: int (optional), status code to match or retry the request.

  Returns:
    requests.Response
  """

  def retry_on_error(e):
    return isinstance(e, (SSLError, ReqConnectionError))

  # generates function to see if the request needs to be retried.
  # if param `code` is None, will not retry and directly pass back the response.
  # Otherwise will retry if status code is not matched.
  def retry_on_result_func(code):
    if code is None:
      return lambda _: False

    return lambda resp: not resp or resp.status_code != code

  @retry(stop_max_delay=wait_sec * 1000, wait_fixed=10 * 1000,
         retry_on_exception=retry_on_error,
         retry_on_result=retry_on_result_func(retry_result_code))
  def _send(url, req_gen):
    resp = None
    logging.info("sending request to %s", url)
    try:
      resp = req_gen()
    except Exception as e:
      logging.warning("%s: request with error: %s", url, e)
      raise e
    return resp

  return _send(url, req_gen)

# TODO(jlewi): basic_auth is no longer supported so we could probably
# delete this code path.
def basic_auth_is_ready(url, username, password, wait_min=15):
  get_url = url + "/kflogin"
  post_url = url + "/apikflogin"

  end_time = datetime.datetime.now() + datetime.timedelta(
      minutes=wait_min)

  wait_time = datetime.datetime.now() - end_time
  resp = _send_req(wait_time.seconds, get_url, lambda: requests.request(
      "GET",
      get_url,
      verify=False), retry_result_code=200)

  logging.info("%s: endpoint is ready; response: %s", get_url, resp.text)
  logging.info("%s: testing login API", post_url)

  wait_time = datetime.datetime.now() - end_time
  resp = _send_req(wait_time.seconds, post_url, lambda: requests.post(
      post_url,
      auth=(username, password),
      headers={
          "x-from-login": "true",
      },
      verify=False))
  logging.info("%s: %s", post_url, resp.text)
  if resp.status_code != 205:
    logging.error("%s: login is failed", post_url)
    return False

  logging.info("%s: testing cookies credentials", url)
  cookie = None
  for c in resp.cookies:
    if c.name == COOKIE_NAME:
      cookie = c
      break
  if cookie is None:
    logging.error("%s: auth cookie cannot be found; name: %s",
                  post_url, COOKIE_NAME)
    return False

  wait_time = datetime.datetime.now() - end_time
  resp = _send_req(wait_time.seconds, url, lambda: requests.get(
      url,
      cookies={
          cookie.name: cookie.value,
      },
      verify=False))
  logging.info("%s: %s", url, resp.status_code)
  logging.info(resp.content)
  return resp.status_code == 200
