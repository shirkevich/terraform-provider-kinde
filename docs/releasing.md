# Release Process

This provider is released from GitHub tags. The OpenTofu registry entry for
`registry.opentofu.org/nxt-fwd/kinde` already exists, so routine provider
versions are published by creating a GitHub release with signed checksums and
then waiting for the OpenTofu registry to index it.

## Current Release Automation

- `.github/workflows/release.yml` runs on tags matching `v*`.
- `.goreleaser.yml` builds provider archives for the supported platforms.
- GoReleaser uploads the platform `.zip` files, `SHA256SUMS`, and
  `SHA256SUMS.sig` to the GitHub release.
- The workflow signs the checksum file with the GPG key from repository secrets:
  `GPG_PRIVATE_KEY` and `PASSPHRASE`.
- `terraform-registry-manifest.json` declares provider protocol metadata.

The OpenTofu registry indexes GitHub release assets and serves them through the
provider registry API. New provider or signing-key submissions must still go
through the OpenTofu registry issue form UI, but normal version releases do not
need a registry PR.

## Prerequisites

Before cutting a release:

1. Make sure `main` contains the commit to release and CI is green.
2. Confirm repository secrets `GPG_PRIVATE_KEY` and `PASSPHRASE` are present.
3. Confirm the signing key is already registered with the OpenTofu registry.
4. Choose the next semantic version.

The latest published OpenTofu registry versions can be checked with:

```sh
curl -fsSL https://registry.opentofu.org/v1/providers/nxt-fwd/kinde/versions
```

## Cut a Release

Use a release tag without the leading `v` in file names and with the leading
`v` in Git tags. For example, the first stable release after `v0.1.0-rc2` is
`v0.1.0`.

```sh
git fetch origin main --tags
git switch main
git pull --ff-only origin main

git tag v0.1.0
git push origin v0.1.0
```

Pushing the tag starts the `Release` GitHub Actions workflow. Wait for it to
complete before checking the registry.

## Verify the GitHub Release

Check that the release exists and includes:

- one `.zip` asset per supported platform
- `terraform-provider-kinde_<version>_SHA256SUMS`
- `terraform-provider-kinde_<version>_SHA256SUMS.sig`

Example:

```sh
gh release view v0.1.0 --repo nxt-fwd/terraform-provider-kinde
gh run list --repo nxt-fwd/terraform-provider-kinde --workflow Release --limit 5
```

## Verify OpenTofu Registry Publication

The registry may lag behind the GitHub release. Poll the versions endpoint until
the new version appears:

```sh
curl -fsSL https://registry.opentofu.org/v1/providers/nxt-fwd/kinde/versions \
  | jq '.versions[].version'
```

Then verify a platform download response:

```sh
curl -fsSL \
  https://registry.opentofu.org/v1/providers/nxt-fwd/kinde/0.1.0/download/linux/amd64 \
  | jq '{version: "0.1.0", os, arch, filename, download_url, shasums_url, shasums_signature_url}'
```

Finally, test installation from a clean OpenTofu config without local provider
overrides:

```hcl
terraform {
  required_providers {
    kinde = {
      source  = "registry.opentofu.org/nxt-fwd/kinde"
      version = "0.1.0"
    }
  }
}
```

```sh
touch /tmp/empty.tfrc
TF_CLI_CONFIG_FILE=/tmp/empty.tfrc tofu init -upgrade
tofu providers
```

If `~/.tofurc` has a `dev_overrides` block for this provider, temporarily move
it aside or run with an empty `TF_CLI_CONFIG_FILE` so the test uses the
registry.

## Failure Handling

- If the workflow fails before publishing assets, fix the issue and push a new
  tag. Do not reuse a public release tag.
- If the workflow creates incomplete assets, delete the GitHub release and tag
  only if it has not been consumed publicly. Prefer a new patch or prerelease
  tag once anything may have been downloaded.
- If the GitHub release is complete but OpenTofu does not index it, confirm the
  tag is semver-compatible, the checksum signature exists, and the signing key is
  registered. If those are correct, open an issue in `opentofu/registry`.
