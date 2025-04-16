#!/usr/bin/env bash
# creates a fake bin directory with mocks for kit and aws
mock_bin() {
  mkdir -p "$BATS_TEST_DIRNAME/bin"

  # mock kit
  cat >"$BATS_TEST_DIRNAME/bin/kit" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == "version" ]]; then
  echo "kit version 1.2.3"
  exit 0
elif [[ "$1" == "login" ]]; then
  echo "[MOCK kit] $@"
  exit 0
elif [[ "$1" == "unpack" ]]; then
  echo "[MOCK kit] $@"
  exit 0
else
  echo "[MOCK kit] $@"
  exit 0
fi
EOF
  chmod +x "$BATS_TEST_DIRNAME/bin/kit"

  # mock aws
  cat >"$BATS_TEST_DIRNAME/bin/aws" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == "ecr" && "$2" == "get-login-password" ]]; then
  echo "[MOCK aws] output"
  exit 0
else
  echo "[MOCK aws] $@"
  exit 0
fi
EOF
  chmod +x "$BATS_TEST_DIRNAME/bin/aws"

  export PATH="$BATS_TEST_DIRNAME/bin:$PATH"
}