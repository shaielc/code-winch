# Sandbox and security model

## 1. Scope, assets, and actors

Coding agents consume untrusted repository content and can execute generated
commands. Repository instructions may attempt prompt injection, secret theft,
network exfiltration, persistence, or sandbox escape. Harness output may contain
malicious ANSI/HTML/URLs. A user may also accidentally grant excessive access.

The first deployment protects:

- repository source, history, uncommitted changes, and workspace availability;
- user input, harness output, canonical events, artifacts, and workflow state;
- credential values and secret-provider references;
- control-plane identities, sessions, policy, audit records, and signing keys;
- runner hosts, other runs, storage, and allowed upstream services; and
- service availability and the integrity/ordering of commands and events.

Actors are authenticated users, workspace/deployment administrators, control
plane and runner machine identities, harness/provider services, and external
services on allowed networks. Repository authors, dependency publishers,
harness output, compromised providers, and unauthenticated network clients are
untrusted. Administrators are trusted to configure policy and infrastructure,
but mistakes and stolen administrator sessions remain in scope. Denial of
service by an authorized user is mitigated, not eliminated, in the first
deployment.

## 2. Trust boundaries

Trust boundaries are: browser to control plane, control plane to runner, runner
to sandbox, sandbox to host/workspace/network, renderer to browser, and all
components to the secret provider. Isolation reduces risk but does not make an
agent trustworthy.

| Boundary | Data crossing it | Required control |
|---|---|---|
| Browser -> control plane | sessions, commands, input, rendered events | TLS, authentication/authorization, CSRF/origin and size/rate checks |
| Control plane -> runner | versioned commands, lease tokens, observations | separate machine identity, authenticated transport, replay and lease fencing |
| Runner -> sandbox | launch specification, scoped mounts and credentials | named policy, capability validation, resource limits, cleanup |
| Sandbox -> repository/host | file and process operations | approved root, traversal checks, disposable workspace; no implied isolation for `local-trusted` |
| Sandbox -> network | DNS and outbound requests | deny by default, explicit allowlist, metadata/link-local block |
| Components -> secret provider | opaque references and scoped values | least privilege, value never persisted, audited use, temporary injection |
| Event pipeline -> renderer -> browser | untrusted text, ANSI, Markdown, URLs | schema/size bounds, sanitization, CSP, no raw DOM HTML or credentials |

The metadata/event store and telemetry exporter are also trust crossings:
authorization, encryption, sensitivity policy, redaction, and bounded labels
apply before data is persisted or exported.

## 3. Threat and mitigation register

The owner is accountable for accepting the residual risk and completing the
mitigation. This register states the control that is intended, not whether it
exists; `docs/state.md` records which of these hold at HEAD.

| ID | Surface and abuse case | Mitigation and verification | Follow-up / residual owner |
|---|---|---|---|
| T01 | Browser: session theft, CSRF, ID guessing, or WebSocket hijack controls another user's run. | Secure cookies or short-lived tokens; authorize every object; validate CSRF and WebSocket origin; reauthorize streams; rate/size limits. | control-plane owner |
| T02 | Browser: a user is misled about effective isolation, credentials, or network access. | Display resolved capabilities before launch; reject rather than silently weaken policy; label `local-trusted` as unisolated. | product/security owner |
| T03 | Control plane: duplicate, stale, or reordered input/events corrupt durable history. | One writer, idempotency keys, transactional outbox, gap-free sequence, lease fencing, replay tests. | application owner |
| T04 | Control plane/storage: unauthorized reads, exports, deletion, or sensitive log/trace labels leak content. | Object authorization; sensitivity policy; content-free audit/telemetry; stable errors that do not reveal existence; canary tests. | control-plane owner |
| T05 | Runner: forged commands/events, a stale runner, or a lost host assumes execution ownership. | Authenticated, rotated runner identity; bounded versioned protocol; command IDs; lease epochs/tokens; reconciliation and revocation. | runner owner |
| T06 | Runner/host: hostile child processes exhaust resources, survive stop, or access host sockets/devices. | Ownership labels, stop escalation, orphan cleanup; container limits; no privileged mode, host namespaces, devices, or Docker socket. | runner owner |
| T07 | Sandbox: repository instructions or generated commands escape, persist, or attack another run. | Disposable per-run workspace; least-privilege named profiles; rootless non-root container baseline; stronger driver required for hostile multi-tenancy. | sandbox/security owner; **accepted residual:** containers and the local driver are not hard hostile multi-tenant boundaries |
| T08 | Repository: traversal, symlink, submodule, or malicious archive accesses files outside the approved root. | Canonicalize beneath approved root; reject traversal/symlink escapes; safe extraction; mount only after policy evaluation; adversarial tests. | sandbox owner |
| T09 | Network: a harness exfiltrates data, reaches cloud metadata/internal services, or abuses DNS/rebinding. | Deny egress by default; dedicated network; domain/port policy plus resolved-IP checks; block link-local/metadata; content-free audit facts. | network/security owner; **accepted residual:** DNS/IP/SNI controls cannot fully identify end-to-end TLS content |
| T10 | Secrets: plaintext credentials enter database, events, artifacts, logs, process environment, or browser. | Store opaque references only; scoped short-lived injection; avoid environment variables; redact before persistence; cleanup; secret canaries. | security owner |
| T11 | Harness/parser: malformed, oversized, or adversarial provider output crashes parsing or bypasses approval semantics. | Bounded incremental parsing; schema validation/fuzzing; structured approval; raw terminal input separately authorized; safe diagnostic fallback. | adapter owner |
| T12 | Renderer: ANSI/Markdown/HTML/URL content executes script, spoofs UI, fetches data, or consumes excessive resources. | Interpret rather than inject output; sanitize; restrictive CSP; safe links; bounded projections; isolate untrusted server renderers. | web/renderer owner |
| T13 | Supply chain: compromised dependencies or images execute in control plane/runner. | Pin image digest and provenance; scan dependencies/images; produce SBOM; restrict adapter and renderer loading. | shared-deployment launch blocker LB08 — release/security owner |
| T14 | Lifecycle: retention worker, export, deletion, backup, or renderer cache leaves unauthorized copies. | One sensitivity policy across stores; authorized integrity-manifested exports; retryable deletion ledger; cache invalidation; backup expiry verification. | data owner |

## 4. Security profiles

Every run selects a named, administrator-approved profile rather than arbitrary
low-level flags.

| Profile | Intended use | Filesystem | Network | Credentials |
|---|---|---|---|---|
| `local-trusted` | developer-only initial use | host workspace | host network | explicit references |
| `container-standard` | normal agent run | dedicated writable copy/worktree | deny by default; allowlist opt-in | scoped injection |
| `container-readonly` | review/analysis | read-only workspace + scratch tmp | deny by default | none by default |

The UI must show that `local-trusted` is not a sandbox. Policies may prohibit it
in shared deployments.

## 5. Data handling and retention defaults

Classification applies to event payloads, input, artifacts, extensions,
renderer caches, exports, logs, traces, metrics, and backups. Metadata needed to
locate or authorize content inherits that content's class. Aggregates use the
most restrictive member. `secret` is a handling prohibition for durable run
data, not permission to persist a credential value.

Durations start when a run reaches a terminal state; active-run data remains
until completion or explicit authorized deletion. These are maximum defaults
for a shared deployment, configurable only to shorter periods until a reviewed
policy feature exists. Legal holds and longer retention are not implicit: they
require documented authorization, scope, expiry, and audit. Backup expiry must
not exceed the class maximum; deletion may be cryptographic/expiry-based for
immutable backups and must report that fact.

| Sensitivity | Examples | Primary/event and artifact retention | Export default | Telemetry and display default | Deletion default |
|---|---|---|---|---|---|
| `public` | published template, explicitly public diagnostic | 365 days | Included in authorized export | Content may be displayed; telemetry still uses IDs/counts, not content | Purge within 30 days of request, including caches |
| `operational` | lifecycle kind, timestamps, status, usage count, resource IDs | 90 days; security audit facts 365 days | Included as manifest/metadata for an authorized scope | IDs, bounded enums, timings, counts allowed; no repository paths, commands, URLs, or free-form values | Purge within 30 days except minimal tombstone/audit fact until its 365-day limit |
| `user-content` | prompts, replies, terminal stream, ordinary diffs | 30 days | Included only for an authorized content export | Authorized UI only; excluded from logs, traces, metrics, and error details | Purge within 7 days; invalidate projections and derived caches |
| `confidential` | private source/artifacts, provider extensions, suspected credential-bearing output | 7 days | Excluded unless separately and explicitly selected by an authorized user; warn and audit | Authorized UI with restrictive rendering; never telemetry content | Purge within 24 hours; remove derived copies and record content-free outcome |
| `secret` | tokens, passwords, private keys, injected credential files | **Never persist in events, artifacts, caches, or telemetry.** Opaque provider reference metadata is `confidential`. | Never export values | Never display after entry or emit values; redact/drop and raise a content-free security signal | Revoke/delete provider value and temporary material immediately; retry cleanup and escalate failures |

Exports require fresh authorization for the workspace and every included
resource, a sensitivity summary, explicit confirmation for `confidential`
data, and a manifest containing IDs, schema versions, sizes, and digests.
Secrets and unauthorized extension fields are always omitted. Export creation
and download are audited without filenames or content. Export bundles inherit
the highest included sensitivity, expire after 24 hours, and are deleted after
successful download when the storage adapter supports it.

Deletion is idempotent and deny-by-default when ownership is ambiguous. It
removes primary objects, blobs, indexes, outbox payloads, projections, renderer
caches, and staged exports, then leaves only a content-free tombstone recording
resource ID, request/completion time, policy version, and partial failures.
Credential deletion means revocation at the provider plus removal of the
reference. Partial deletion is retried and visible to operators; success is not
claimed until every mutable store acknowledges it and immutable-copy expiry is
recorded. Implementations are verified under fake time and crash replay.

### Telemetry, errors, and audit fields

Default logs and traces may contain generated resource IDs, stable error code,
component, operation, bounded status, duration, retry count, size, and sequence
number. Metrics use bounded labels only. They must not contain input/output,
commands, diffs, repository paths, filenames, branches, URLs/query strings,
headers, provider payloads/extensions, rendered text, secret references, or
credential values. Free-form fields are content and therefore prohibited.

Public errors have a stable code, safe summary, retryability, correlation ID,
and an actionable next step. Detailed causes remain in access-controlled
content-safe diagnostics. Authorization failures do not distinguish absent
from forbidden resources. Redaction occurs before persistence or export, and a
redaction failure drops/quarantines the field rather than logging it. API error
mapping, the audit trail, secret canaries, and telemetry enforcement are each
bound by these rules.

## 6. Container baseline

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

## 7. Credentials and login

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

## 8. Network and communication proxy

A communication proxy can enforce domain/port allowlists, attach scoped
credentials to approved upstream requests, log metadata, cap bandwidth, and
block link-local/cloud-metadata endpoints. TLS interception is not assumed; for
end-to-end TLS, enforcement is based on DNS/IP/SNI with acknowledged limits.

Provider API proxying is an optional adapter capability, not embedded in the
event renderer. It is useful when structured agent communication, billing,
redaction, or credential non-disclosure requires the harness to contact a local
endpoint. The proxy emits audit facts but does not persist sensitive bodies by
default.

## 9. Application security

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

## 10. Approval and policy model

Adapters normalize provider requests into explicit approval events. A policy
engine can auto-deny, require a user, or allow narrowly described operations.
Approval displays the exact command/tool, working directory, file/network
scope, and whether it exceeds the active profile. Approval IDs are single-use,
expire, bind to the exact operation digest, and are audited.

Top-level workflows cannot silently broaden run permissions. Effective policy
is the intersection of deployment, user/workspace, workflow, and sandbox
profile permissions.

## 11. Shared-deployment launch blockers

The release/security owner records evidence and signs off every item. An item
cannot be waived silently; a time-bounded exception names an accountable owner,
expiry, affected deployment, compensating control, and rollback trigger.

| ID | Blocker and evidence | Owner |
|---|---|---|
| LB01 | Authentication, object authorization, CSRF/origin, session, and API/WebSocket rate/size tests pass. | Control-plane owner |
| LB02 | `local-trusted` is prohibited by shared-deployment policy and shown as unisolated wherever enabled for development. | Product/security owner |
| LB03 | Selected sandbox driver has documented threat review; traversal/extraction, resource exhaustion, forced stop, crash, and orphan-cleanup tests pass. | Sandbox/runner owner |
| LB04 | Egress deny/allowlist, DNS, link-local, and cloud-metadata tests pass on the deployment network. | Network/security owner |
| LB05 | Credential flow uses opaque references and scoped injection; canary values occur zero times in events, artifacts, exports, default logs, and traces. | Security owner |
| LB06 | Renderer/parser schema, fuzz, malicious ANSI/HTML/Markdown/URL, CSP, and bounded-resource tests pass. | Web/adapter owner |
| LB07 | Retention, authorized export, deletion retry, cache invalidation, and backup-expiry tests pass for every sensitivity class. | Data owner |
| LB08 | Images/dependencies are scanned, critical findings resolved or explicitly excepted, pinned image provenance recorded, and an SBOM produced. | Release/security owner |
| LB09 | Incident response names on-call contacts and exercises credential revocation, runner revocation, audit preservation, containment, and user notification. | Operations/security owner |
| LB10 | Runner identity rotation, lease fencing, stale-event rejection, daemon recovery, restore, and emergency revocation are exercised. | Runner/operations owner |

Hostile public or mutually untrusted tenants are out of scope for the container
deployment. Enabling them is itself a launch blocker until a stronger isolation
driver and separate-runner-host design receive threat review.

## 12. Security validation checklist

Before a shared deployment:

- threat-model each sandbox driver and credential flow;
- test path/symlink traversal and archive extraction;
- test resource exhaustion, stop escalation, daemon crash, and orphan cleanup;
- verify egress restrictions including DNS and metadata endpoints;
- fuzz harness parsers, ANSI handling, schemas, and API payloads;
- scan images/dependencies and produce an SBOM;
- exercise secret canaries to confirm logs/events/artifacts are redacted; and
- document incident response and emergency runner revocation.
