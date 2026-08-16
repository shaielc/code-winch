# Supply-chain scanning and SBOMs

Pull requests run `make scan` and fail on unsuppressed high or critical
dependency or image vulnerabilities. The scan covers the Go module graph, the
web workspace, and the `winchd` and `web` images. A scanner that is missing,
cannot update its advisory database, or otherwise exits unsuccessfully fails
the job; it is never treated as a clean report.

`make sbom` builds both images and writes CycloneDX JSON documents for each
image and the repository to `sbom/`. CI validates that each document has
components, publishes them as the `sbom` workflow artifact, and attaches the
same files to a published release. Image SBOMs include `baseImageDigest`, copied
from the final pinned `FROM` instruction, so release evidence can be matched to
the deployment source.

## Triage and exceptions

1. Re-run the job once to distinguish an advisory-database or registry outage
   from a repeatable finding. Never merge by disabling or bypassing the scan.
2. For a finding, identify the package, advisory, reachable use, and available
   fixed version. Upgrade or remove the dependency whenever possible.
3. If risk must temporarily be accepted, add its exact advisory ID to
   `security/suppressions.yaml`. Every entry must contain non-empty `owner` and
   `reason` fields and an ISO `expires` date, for example:

   ```json
   {
     "id": "CVE-2099-1234",
     "owner": "security@example.com",
     "reason": "Not reachable; upgrade is being validated",
     "expires": "2099-02-01"
   }
   ```

   `make suppressions-check` rejects malformed records and dates earlier than
   the current UTC date before scanning begins. Reviewers should require the
   reason to explain exposure and remediation, not merely repeat the finding.
4. If the scanner or its advisory service remains unavailable, record the
   outage with the release/security owner and wait for recovery. Because tool
   setup, database download, scan, and artifact upload are all fail-closed, an
   outage cannot create release evidence or a passing pull-request check.

The sandbox image is added to the image list when its Dockerfile lands; until
then there is no sandbox image in this repository to build or scan.
