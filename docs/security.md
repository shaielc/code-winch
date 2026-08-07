# Sandbox and security model

## 1. Trust boundaries and threats

Coding agents consume untrusted repository content and can execute generated
commands. Repository instructions may attempt prompt injection, secret theft,
network exfiltration, persistence, or sandbox escape. Harness output may contain
malicious ANSI/HTML/URLs. A user may also accidentally grant excessive access.

Trust boundaries are: browser to control plane, control plane to runner, runner
to sandbox, sandbox to host/workspace/network, renderer to browser, and all
components to the secret provider. Isolation reduces risk but does not make an
agent trustworthy.

## 2. Security profiles

Every run selects a named, administrator-approved profile rather than arbitrary
low-level flags.

| Profile | Intended use | Filesystem | Network | Credentials |
|---|---|---|---|---|
| `local-trusted` | developer-only initial use | host workspace | host network | explicit references |
| `container-standard` | normal agent run | dedicated writable copy/worktree | deny by default; allowlist opt-in | scoped injection |
| `container-readonly` | review/analysis | read-only workspace + scratch tmp | deny by default | none by default |

The UI must show that `local-trusted` is not a sandbox. Policies may prohibit it
in shared deployments.

## 3. Container baseline

The Docker adapter should apply, where supported:

- rootless engine/user namespace and non-root container user;
- pinned image digest and recorded image provenance;
- dropped Linux capabilities, `no-new-privileges`, and a default seccomp
  profile;
- read-only root filesystem with explicit bounded writable mounts;
- CPU, memory, process-count, disk/quota, and wall-clock limits;
- no host Docker socket, host PID/IPC namespace, devices, or privileged mode;
- dedicated per-run network with deny-by-default egress enforcement outside the
  container namespace; and
- deterministic labels enabling reconciliation and orphan cleanup.

Docker configuration alone is not a hard multi-tenant boundary. Hostile public
workloads should use a stronger isolation driver such as a microVM and separate
runner hosts.

Workspace paths are resolved beneath an administrator-approved root, checked
against traversal/symlink escapes, and mounted only after policy evaluation.
Prefer a disposable worktree/copy. Changes return as reviewed diffs/artifacts
rather than automatically mutating the user's primary checkout.

## 4. Credentials and login

The domain stores credential metadata and an opaque secret-provider reference;
plaintext tokens never enter the database or event log. Login patterns are:

1. **host-mediated OAuth/device flow:** control plane exposes the provider URL
   and code, stores the resulting token in a secret provider, and injects a
   scoped token at launch;
2. **user-supplied API token:** accepted over TLS, immediately stored, then
   referenced by ID; and
3. **interactive harness login:** only when unavoidable, proxied as a sensitive
   structured exchange and disabled by default for raw terminal sessions.

Prefer short-lived, least-privilege credentials scoped to a provider and
workspace. Mount a temporary credential file or use a narrowly scoped broker;
avoid environment variables when the harness supports safer mechanisms. Delete
temporary material during cleanup, redact known secrets before persistence, and
never return credential values to the browser after entry.

Do not mount a developer's entire home directory, SSH directory, cloud config,
or agent socket into a sandbox. Git signing and remote push are separate,
explicit capabilities.

## 5. Network and communication proxy

A communication proxy can enforce domain/port allowlists, attach scoped
credentials to approved upstream requests, log metadata, cap bandwidth, and
block link-local/cloud-metadata endpoints. TLS interception is not assumed; for
end-to-end TLS, enforcement is based on DNS/IP/SNI with acknowledged limits.

Provider API proxying is an optional adapter capability, not embedded in the
event renderer. It is useful when structured agent communication, billing,
redaction, or credential non-disclosure requires the harness to contact a local
endpoint. The proxy emits audit facts but does not persist sensitive bodies by
default.

## 6. Application security

- Authenticate browser and runner identities separately. Remote runners use
  mutually authenticated, rotated machine credentials.
- Authorize every workspace, run, artifact, input, credential, and workflow
  operation; possession of an ID is insufficient.
- Use secure, HTTP-only, same-site cookies with CSRF protection or short-lived
  bearer tokens appropriate to the deployment.
- Validate WebSocket origin and reauthorization; impose message/rate limits.
- Sanitize Markdown/HTML, safely interpret ANSI, use a restrictive CSP, and do
  not render provider output via raw DOM HTML.
- Record security-relevant audit events: policy changes, credential use (not
  value), run launch/stop, approvals, exports, and administrative actions.
- Encrypt transport and storage, define content retention/deletion policies,
  and exclude content from telemetry by default.

## 7. Approval and policy model

Adapters normalize provider requests into explicit approval events. A policy
engine can auto-deny, require a user, or allow narrowly described operations.
Approval displays the exact command/tool, working directory, file/network
scope, and whether it exceeds the active profile. Approval IDs are single-use,
expire, bind to the exact operation digest, and are audited.

Top-level workflows cannot silently broaden run permissions. Effective policy
is the intersection of deployment, user/workspace, workflow, and sandbox
profile permissions.

## 8. Security validation checklist

Before a shared deployment:

- threat-model each sandbox driver and credential flow;
- test path/symlink traversal and archive extraction;
- test resource exhaustion, stop escalation, daemon crash, and orphan cleanup;
- verify egress restrictions including DNS and metadata endpoints;
- fuzz harness parsers, ANSI handling, schemas, and API payloads;
- scan images/dependencies and produce an SBOM;
- exercise secret canaries to confirm logs/events/artifacts are redacted; and
- document incident response and emergency runner revocation.
