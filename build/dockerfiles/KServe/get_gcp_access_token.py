#!/usr/bin/env python3

import google.auth
import google.auth.transport.requests
import sys


def get_access_token():
    try:
        credentials, project = google.auth.default(
            scopes=["https://www.googleapis.com/auth/cloud-platform"]
        )
        request = google.auth.transport.requests.Request()
        credentials.refresh(request)
        return credentials.token
    except Exception as e:
        print(f"Error getting access token: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    print(get_access_token())
