#!/usr/bin/env bats


   load './helpers/mock_bin.sh'
   bats_load_library bats-support
   bats_load_library bats-assert

setup() {
  mock_bin
  rm -rf tmp
  mkdir tmp
}

teardown() {
  rm -rf tmp
}

@test "no args → prints Usage and exits 1" {
  run bash ./entrypoint.sh
  assert_failure
  assert_output --partial "Usage: entrypoint.sh"
}

@test "with KIT_USER & KIT_PASSWORD → calls kit login then unpack" {
  export KIT_USER=testuser
  export KIT_PASSWORD=testpass

  run bash ./entrypoint.sh kit://foo/bar:latest tmp
  assert_success
  assert_output --partial "[MOCK kit] login"
  assert_output --partial "[MOCK kit] unpack foo/bar:latest"
}

@test "with AWS_ROLE_ARN → calls aws ecr login then unpack" {
  unset KIT_USER KIT_PASSWORD
  export AWS_ROLE_ARN=arn:aws:iam::123456789012:role/testrole
  export AWS_ECR_REGION=us-west-2

  run bash ./entrypoint.sh kit://myrepo tmp
  assert_success
  assert_output --partial [MOCK kit] login --username AWS --password-stdin "[MOCK aws] output"
  assert_output --partial "[MOCK kit] unpack myrepo"
}

@test "unpacks with KIT_UNPACK_FLAGS" {
  export KIT_UNPACK_FLAGS="--dry-run --verbose"
  run bash ./entrypoint.sh kit://foo tmp
  assert_success
  assert_output --partial "Unpack options: --dry-run --verbose"
}

@test "no credentials set → only unpack, no login" {
  unset KIT_USER KIT_PASSWORD AWS_ROLE_ARN AWS_ECR_REGION

  run bash entrypoint.sh kit://foo/bar tmp
  assert_success
  refute_output --partial "login"        # no kit login
  refute_output --partial "get-login-password"  # no aws login
  assert_output --partial "[MOCK kit] unpack foo/bar"
}

@test "KIT_USER/PASSWORD takes precedence over AWS_ROLE_ARN" {
  export KIT_USER=foo
  export KIT_PASSWORD=bar
  export AWS_ROLE_ARN=arn:aws:iam::123456789012:role/whatever
  export AWS_ECR_REGION=us-east-1

  run bash entrypoint.sh kit://repo tmp
  assert_success
  assert_output --partial "[MOCK kit] login"            # kit login happened
  refute_output --partial "get-login-password"          # aws login did NOT happen
}

@test "parses multi‑flag KIT_UNPACK_FLAGS into unpack args" {
  export KIT_UNPACK_FLAGS="--foo bar --baz qux"
  run bash entrypoint.sh kit://foo tmp
  assert_success
  # the mock prints its args so we can verify both flags are passed
  assert_output --partial "[MOCK kit] unpack foo -d tmp --foo bar --baz qux"
}

@test "repo name strips only the kit:// prefix" {
  run bash entrypoint.sh kit://registry.local:5000/ns/proj:123 tmp
  assert_success
  assert_output --partial "[MOCK kit] unpack registry.local:5000/ns/proj:123 -d tmp"
}