# Release Process

This document outlines the steps to release a new version of autotidy.

## Version Locations

Most version numbers are derived automatically from git tags. Only these files contain hardcoded versions:

| File | Line | Notes |
|------|------|-------|
| `flake.nix` | `version = "X.Y.Z"` | Update before tagging |
| Homebrew formula | `url` and `sha256` | Update after release is published |

**Automatically derived:**
- `Makefile` uses `git describe --tags`
- `nfpm.yaml` version is injected by the release workflow
- Go binary version is set via `-ldflags` at build time

## Release Steps

### 1. Update flake.nix version

Edit `flake.nix` and update the version string:

```nix
version = "X.Y.Z";
```

### 2. Commit the version bump

```bash
git add flake.nix
git commit -m "Bump version to X.Y.Z"
```

### 3. Create and push the tag

```bash
git tag vX.Y.Z
git push origin master
git push origin vX.Y.Z
```

This triggers the GitHub Actions release workflow.

### 4. Wait for the release workflow to complete

The workflow will:
- Build binaries for all platforms (darwin/linux arm64+amd64, windows amd64)
- Create archives (.tar.gz for macOS/Linux, .zip for Windows)
- Build .deb and .rpm packages for Linux
- Generate `checksums.txt` with SHA256 hashes
- Create a GitHub Release with all artifacts

### 5. Update Homebrew formula

Once the release is published, update the Homebrew tap:

1. Go to the GitHub Release page
2. Download `checksums.txt` or copy the SHA256 values for:
   - `autotidy_X.Y.Z_darwin_arm64.tar.gz`
   - `autotidy_X.Y.Z_darwin_amd64.tar.gz`

3. Update the formula in `homebrew-tap/Formula/autotidy.rb`:

```ruby
class Autotidy < Formula
  desc "Automatically watch and organize files using configurable rules"
  homepage "https://github.com/prettymuchbryce/autotidy"
  version "X.Y.Z"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/prettymuchbryce/autotidy/releases/download/vX.Y.Z/autotidy_X.Y.Z_darwin_arm64.tar.gz"
      sha256 "PASTE_ARM64_SHA256_HERE"
    else
      url "https://github.com/prettymuchbryce/autotidy/releases/download/vX.Y.Z/autotidy_X.Y.Z_darwin_amd64.tar.gz"
      sha256 "PASTE_AMD64_SHA256_HERE"
    end
  end

  # ... rest of formula
end
```

4. Commit and push the formula update:

```bash
cd /path/to/homebrew-tap
git add Formula/autotidy.rb
git commit -m "autotidy X.Y.Z"
git push
```

### 6. Verify installation

```bash
brew update
brew upgrade autotidy
# or for new installs:
brew install prettymuchbryce/tap/autotidy
```

## Checklist

- [ ] Update version in `flake.nix`
- [ ] Commit version bump
- [ ] Create and push tag
- [ ] Wait for GitHub Actions to complete
- [ ] Copy SHA256 values from release checksums.txt
- [ ] Update Homebrew formula with new URLs and SHA256 values
- [ ] Push Homebrew formula update
- [ ] Test installation with `brew install`
