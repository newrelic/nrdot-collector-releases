---
name: version-bump-analysis
description: |
  Analyzes OpenTelemetry Collector component version bumps and compiles a markdown
  summary of relevant changes (breaking changes, deprecations, bug fixes) from the
  upstream CHANGELOG files. Only includes changes for components used in the nrdot
  releases repo. Outputs a formatted comment ready to paste into a GitHub PR.
trigger: /version-bump-analysis
type: agent
parameters:
  - name: pr_url
    description: "GitHub PR URL containing the version bump (e.g., https://github.com/newrelic/nrdot-collector-releases/pull/488)"
    required: true
permissions:
  - Read
  - Glob
  - Bash
---

# Version Bump Analysis Skill

Analyzes OpenTelemetry Collector component version bumps and compiles a markdown summary of relevant changes for PR comments.

## Input

- A GitHub PR URL from the `newrelic/nrdot-collector-releases` repository
- The PR will contain version bumps with a title like "feat: Bump OTEL beta core to v0.144.0"
- The PR description will include version ranges like "v0.142.0...v0.144.0"

## Output Format

Generate a markdown comment that can be pasted directly into the PR, following this structure:

```markdown
### Potentially Relevant changes

#### Contrib v[VERSION]

##### 🛑 Breaking changes 🛑

- [List breaking changes with full descriptions from CHANGELOG]

##### 🚩 Deprecations 🚩

- [List deprecations with full descriptions from CHANGELOG]

##### 🧰 Bug fixes 🧰

- [List bug fixes that may be relevant]

#### Core v[VERSION]

[Same structure as Contrib if there are relevant changes]
```

## ⚠️ IMPORTANT: Use Bash Script for CHANGELOG Files

**Do NOT fetch full CHANGELOG files directly** - they are very large (~66k tokens) and inefficient. Instead, use the provided bash script to extract only the relevant version section:

```bash
.claude/skills/version-bump-analysis/extract-changelog-version.sh <changelog_url> <version>
```

This script efficiently fetches and extracts only the relevant version section using curl and awk.

### Example Usage

```bash
# Extract v0.144.0 from OTel Core CHANGELOG
.claude/skills/version-bump-analysis/extract-changelog-version.sh \
  "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/main/CHANGELOG.md" \
  "v0.144.0"

# Extract v0.144.0 from OTel Contrib CHANGELOG
.claude/skills/version-bump-analysis/extract-changelog-version.sh \
  "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/main/CHANGELOG.md" \
  "v0.144.0"
```

## Implementation Steps

1. **Extract list of used components from manifest files**:
   - Use Glob to find all `distributions/*/manifest.yaml` files
   - Use Read to load each manifest file
   - Parse the YAML to extract component names from:
     - `receivers:` section
     - `processors:` section
     - `exporters:` section
     - `connectors:` section
     - `extensions:` section
     - `providers:` section
   - Extract component names from gomod paths (e.g., `receiver/filelogreceiver` → `receiver/filelog`)
   - Build a set of all unique component names used across all distributions
   - Component names in CHANGELOG use format: `receiver/filelog`, `processor/cumulativetodelta`, etc.

2. **Extract version information from the PR**:
   - Use `gh pr view <pr_number> --json body --jq .body` to fetch the PR description
   - Extract the "from" and "to" versions from the PR description
   - The description includes links like: `Beta Core: [v0.142.0...v0.144.0]` and `Beta Contrib: [v0.142.0...v0.144.0]`

3. **Fetch CHANGELOG sections for each version**:
   - Contrib CHANGELOG: `https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/main/CHANGELOG.md`
   - Core CHANGELOG: `https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/main/CHANGELOG.md`
   - **Use the extract-changelog-version.sh script** to efficiently extract only the relevant version sections
   - For each version in the range, call: `.claude/skills/version-bump-analysis/extract-changelog-version.sh <url> <version>`
   - This returns only the relevant section for that version (typically <1000 lines vs 66k+ for full file)
   - The script will exit with status 1 if the version is not found, allowing error detection

4. **Parse version sections**:
   - For each version between "from" and "to" (inclusive of "to", exclusive of "from")
   - Extract sections marked as:
     - `### 🛑 Breaking changes 🛑` or `## 🛑 Breaking changes 🛑`
     - `### 🚩 Deprecations 🚩` or `## 🚩 Deprecations 🚩`
     - `### 🧰 Bug fixes 🧰` or `## 🧰 Bug fixes 🧰`
   - Preserve the full descriptions and component names from the CHANGELOG

5. **Filter for component relevance**:
   - Only include CHANGELOG entries that mention components from the manifest files
   - Match component names like `receiver/filelog`, `processor/batch`, `exporter/otlp`, etc.
   - Include entries where the component name appears at the start of the bullet point
   - Format: `` `component/name`: description ``
   - Also include changes to core collector functionality that affect all components (if any)

6. **Format the output**:
   - Use the exact emoji and header format from the example
   - Preserve markdown formatting from the CHANGELOG
   - Include component names in the format `` `component/name`: description ``
   - Add clear section headers for Contrib and Core
   - Only include sections that have content (omit empty sections)
   - If no relevant changes found, output: "No breaking changes, deprecations, or relevant bug fixes found for components used in this repo."

7. **Save the output to a file**:
   - Create the `.tmp` directory if it doesn't exist using: `mkdir -p .tmp`
   - Write the markdown output to a file in `.tmp` directory with naming pattern: `version-bump-analysis-pr{PR_NUMBER}.md`
   - Example: For PR #488, write to `.tmp/version-bump-analysis-pr488.md`
   - Use the Write tool to save the formatted markdown content to this file

## Important Notes

- The skill should handle multiple version bumps (e.g., v0.142.0 to v0.144.0 means analyzing both v0.143.0 and v0.144.0)
- Preserve the exact wording from the CHANGELOG - don't summarize or paraphrase
- If there are no breaking changes or deprecations for a section, omit that section entirely
- The output should be ready to copy-paste directly into a GitHub PR comment
- Always include the version number in section headers (e.g., "Contrib v0.144.0")
- If bumping multiple versions, combine changes from all versions into a single output
- **CRITICAL**: Only include changes for components that are listed in the manifest.yaml files

## Component Name Mapping

When matching CHANGELOG entries to manifest components:
- Manifest: `receiver/filelogreceiver` → CHANGELOG: `receiver/filelog`
- Manifest: `processor/batchprocessor` → CHANGELOG: `processor/batch`
- Manifest: `exporter/otlpexporter` → CHANGELOG: `exporter/otlp`
- Pattern: Remove the trailing "receiver", "processor", "exporter", etc. suffix to get the CHANGELOG name

## Example Reference

See https://github.com/newrelic/nrdot-collector-releases/pull/464#issuecomment-3671968859 for the expected output format.
