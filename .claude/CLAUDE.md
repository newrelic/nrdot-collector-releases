# NRDOT Collector Releases Project Instructions

## Distribution Changes

When making changes to any distribution manifest (`distributions/*/manifest.yaml`):

1. **Always run `make licenses`** after modifying manifest files
   - This will automatically regenerate sources and update THIRD_PARTY_NOTICES.md
2. Include both the manifest.yaml and THIRD_PARTY_NOTICES.md changes in your commit

## Build Commands

- `make licenses` - Update THIRD_PARTY_NOTICES.md (also runs generate-sources as dependency)
- `make generate-sources` - Generate collector sources from manifest files only
- `make build` - Build all distributions
- `make ci` - Run full CI checks (manifests-check, build, licenses-check, etc.)
