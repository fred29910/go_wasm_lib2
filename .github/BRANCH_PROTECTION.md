# Branch Protection Setup Guide

This document describes the recommended branch protection rules for the `main` branch.
Configure these in **Settings → Branches → Branch protection rules** on GitHub.

## Recommended Rules for `main`

- **Require a pull request before merging**
  - ✅ Require approvals: **1**
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ Require review from Code Owners

- **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Required status checks**:
    - `lint` — Go vet & mod tidy
    - `test (1.25.1, ubuntu-latest)` — Unit tests
    - `build-wasm-go` — Standard Go WASM build
    - `build-wasm-tinygo` — TinyGo WASM build
    - `generate-sdk` — Petstore SDK generation
    - `lint-ts` — oxlint on generated TypeScript
    - `verify` — Full verification summary

- **Require conversation resolution before merging**
  - ✅ Enabled

- **Require signed commits** (optional)
  - ❌ Disabled (enable for higher security)

- **Include administrators**
  - ✅ Enabled

- **Restrict pushes**
  - ✅ Restrict pushes that match `main`
  - Allow force pushes: ❌
  - Allow deletions: ❌
