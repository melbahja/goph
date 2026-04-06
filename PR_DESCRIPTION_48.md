# fix(auth): support in-memory private key content in Key() (#48)
## Related Issue
Closes #48
## Summary
This PR fixes private key authentication when the private key is provided as in-memory PEM content instead of a filesystem path.
After this change, `goph.Key(...)` accepts both:
- a private key file path (existing behavior),
- raw private key content (new supported behavior).
## Problem
Issue #48 reports that users storing private keys only in memory cannot use:
```go
auth, err := goph.Key(serverData.SshPrivateKey, "")
```
because `Key()` previously treated the first argument strictly as a file path and attempted `ReadFile(...)`, which fails for raw PEM content.
## Root Cause
- `Key()` calls `GetSigner(...)`.
- `GetSigner(...)` always attempted to read `prvFile` from disk.
- Even though `RawKey(...)` existed, users expected `Key(...)` to work with in-memory PEM input directly.
## What Changed
### 1) Auth logic
- Updated `GetSigner(prvFile, passphrase)` in `auth.go`:
  - Detects PEM-style key content (`-----BEGIN ...`).
  - If PEM content is provided, it delegates to `GetSignerForRawKey(...)`.
  - Otherwise, it keeps the original file-path behavior.
### 2) Tests
- Added `auth_test.go` with coverage for:
  - `GetSigner` using raw in-memory private key content.
  - `Key` using raw in-memory private key content.
  - `GetSigner` using private key file path (regression/compat check).
### 3) Documentation
- Updated `README.md` with a new example showing in-memory private key usage:
  - `goph.Key(privateKeyPEM, "")`
## Backward Compatibility
- Backward compatible: **Yes**
- Existing file-path usage of `goph.Key(...)`: **Unchanged**
- Existing passphrase flow: **Unchanged**
## Validation
```bash
go test ./...
```
All tests pass.
## Files Changed
- `auth.go`
- `auth_test.go` (new)
- `README.md`
## Change Type
- [x] Bug fix
- [x] Tests
- [x] Documentation
