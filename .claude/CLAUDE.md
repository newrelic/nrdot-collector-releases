# NRDOT Collector Releases Project Instructions

## Distribution Changes

When making changes to any distribution manifest (`distributions/*/manifest.yaml`):

1. **Always run `make licenses`** after modifying manifest files
   - This will automatically regenerate sources and update THIRD_PARTY_NOTICES.md
2. Include both the manifest.yaml and THIRD_PARTY_NOTICES.md changes in your commit

## Build Commands

- `make licenses` - Update THIRD_PARTY_NOTICES.md (also runs `generate-license-sources` as dependency)
- `make build` - Build all distributions
- `make ci` - Run full CI checks (manifests-check, build, licenses-check, etc.)

## Pushing Branches

Follow the project git conventions:

- Branch: `$developer\<tbranch-name>` `$developer` is the current git user (e.g., `$(git config user.email | cut -d@ -f1))`)
- Message: Single line with conventional commit prefix (`feat`, `fix` etc.)

```bash
developer="$(git config user.email | cut -d@ -f1)"
branch="$developer/<title>"
git checkout -b "$branch"
git add <files>
git commit -m "<prefix>: <message>"
git push -u origin "$branch"
```
