"""Render the control-panel page; buttons call the HTTP API."""

import html
from typing import Any
from urllib.parse import quote, urlparse

from .integrations.codex import canonical_task_url
from .integrations.repository import task_branch

STATUS_ORDER = ["in_progress", "blocked", "pending", "completed"]
def rows(tracker: dict[str, Any], state: dict[str, Any]) -> list[dict[str, Any]]:
    overrides = state["tasks"]
    result = []
    for task in tracker["tasks"]:
        # A retired entry no longer overrides the tracker, but it still carries the links.
        record = overrides.get(task["id"], {})
        lease = record if not record.get("expired") and task["status"] != "completed" and record.get("status") != "completed" else None
        result.append(
            {
                "id": task["id"],
                "title": task["title"],
                "depends_on": task["depends_on"],
                "status": lease["status"] if lease else task["status"],
                "owner": (lease or task).get("owner"),
                "updated_at": record.get("updated_at"),
                "task_url": record.get("task_url"),
                "stages": record.get("stages", {}),
                "pull_request": record.get("pull_request"),
                "pull_request_state": record.get("pull_request_state"),
                "local": bool(lease),
                "prepared": record.get("prepared", False),
                "prepare_error": record.get("prepare_error"),
            }
        )
    return result


def available_ids(entries: list[dict[str, Any]]) -> set[str]:
    done = {entry["id"] for entry in entries if entry["status"] == "completed"}
    return {
        entry["id"]
        for entry in entries
        if entry["status"] == "pending" and set(entry["depends_on"]) <= done
    }


def forest(entries: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Nest every task under the dependency that unlocks it last.

    Dependencies form a graph, not a tree: a task can wait on several others. Hanging it off the
    deepest of them keeps each task in exactly one place, directly below whatever gates it, and
    leaves the dependencies that lost the tie to render as chips on the node. The trailing pass
    picks up whatever the roots never reached, which a dependency cycle would leave behind.
    """
    known = {entry["id"]: entry for entry in entries}
    depth: dict[str, int] = {}

    def rank(task_id: str, walked: frozenset[str] = frozenset()) -> int:
        if task_id in depth:
            return depth[task_id]
        entry = known.get(task_id)
        if entry is None or task_id in walked:
            return 0
        depth[task_id] = 1 + max(
            (rank(other, walked | {task_id}) for other in entry["depends_on"]), default=-1
        )
        return depth[task_id]

    parents: dict[str, str | None] = {}
    children: dict[str | None, list[dict[str, Any]]] = {}
    for entry in entries:
        options = [
            dependency
            for dependency in entry["depends_on"]
            if dependency in known and dependency != entry["id"]
        ]
        parent = max(options, key=rank) if options else None
        parents[entry["id"]] = parent
        children.setdefault(parent, []).append(entry)

    placed: set[str] = set()

    def build(entry: dict[str, Any]) -> dict[str, Any]:
        placed.add(entry["id"])
        waiting = [child for child in children.get(entry["id"], []) if child["id"] not in placed]
        return {
            "entry": entry,
            "waits_on": [
                dependency
                for dependency in entry["depends_on"]
                if dependency != parents[entry["id"]]
            ],
            "children": [build(child) for child in waiting],
        }

    trees = [build(entry) for entry in children.get(None, [])]
    return trees + [build(entry) for entry in entries if entry["id"] not in placed]


PAGE = r"""<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Code Winch scheduler</title>
<style>
  :root {{ color-scheme: light dark; --line: #8883; }}
  body {{ font: 15px/1.6 system-ui, sans-serif; margin: 0 auto; padding: 2.5rem 2rem; max-width: 92rem; }}
  header {{ display: flex; align-items: baseline; gap: 1rem; flex-wrap: wrap; margin-bottom: .4rem; }}
  h1 {{ font-size: 1.3rem; margin: 0; }}
  .sub {{ opacity: .7; font-size: .875rem; margin-bottom: 1.75rem; }}
  .msg {{ padding: .75rem 1rem; border: 1px solid var(--line); border-radius: 8px; margin-bottom: 1.5rem; }}
  /* Scrolling horizontally also clips vertically, so leave the last row's menu room to open. */
  .wrap {{ overflow-x: auto; padding-bottom: 5.5rem; }}
  table {{ border-collapse: collapse; width: 100%; font-size: .9rem; }}
  th, td {{ text-align: left; padding: .8rem .75rem; border-bottom: 1px solid var(--line); white-space: nowrap; }}
  th {{ font-weight: 600; opacity: .65; font-size: .78rem; text-transform: uppercase; letter-spacing: .06em; }}
  tbody tr:hover td {{ background: #8881; }}
  tbody tr {{ scroll-margin-top: 1.5rem; }}
  tbody tr:target td {{ background: #3b82f61f; }}
  td.title {{ white-space: normal; min-width: 15rem; line-height: 1.45; }}
  td.deps {{ white-space: normal; min-width: 15rem; max-width: 19rem; line-height: 1.9; }}
  td.pr a {{ color: inherit; text-decoration: none; }}
  th.stage, td.stage {{ text-align: center; }}
  .stage-icon {{ display: inline-flex; padding: .25rem; border-radius: 6px; cursor: pointer;
    color: inherit; text-decoration: none; }}
  .stage-icon:hover {{ background: #8882; }}
  /* The icon carries the stage's whole state: the default colour means the next click launches,
     green means a conversation is live to open, amber means a submission may or may not have
     been accepted, and nearly invisible means it cannot be launched yet. */
  .stage-icon[data-state="open"] .icon {{ opacity: 1; fill: #16a34a; }}
  .stage-icon[data-state="unknown"] .icon {{ opacity: .9; fill: #d97706; }}
  button.icon-button {{ padding: .25rem; border: 0; background: none; color: #dc2626; }}
  button.icon-button .icon {{ opacity: .85; }}
  button.icon-button:hover {{ background: #dc262622; }}
  button.icon-button:hover .icon {{ opacity: 1; }}
  .pr-icon {{ display: inline-flex; padding: .25rem; border-radius: 6px; color: inherit;
    text-decoration: none; }}
  .pr-icon:not([data-state="blocked"]) {{ cursor: pointer; }}
  .pr-icon:not([data-state="blocked"]):hover {{ background: #8882; }}
  .pr-icon[data-state="found"] .icon {{ opacity: 1; fill: #16a34a; }}
  .pr-icon[data-state="blocked"] {{ cursor: not-allowed; }}
  .pr-icon[data-state="blocked"] .icon {{ opacity: .25; }}
  .stage-cell:not([data-reason=""]) .stage-icon[data-state="launch"] {{ cursor: not-allowed; }}
  .stage-cell:not([data-reason=""]) .stage-icon[data-state="launch"] .icon {{ opacity: .15; }}
  #stage-menu {{ position: absolute; z-index: 5; min-width: 13rem; padding: .35rem;
    background: Canvas; border: 1px solid var(--line); border-radius: 8px;
    box-shadow: 0 2px 10px #0003; display: flex; flex-direction: column; }}
  /* The id selector outranks the hidden attribute's display: none, so restate it. */
  #stage-menu[hidden] {{ display: none; }}
  #stage-menu button, #stage-menu a {{ font: inherit; font-size: .8rem; text-align: left;
    margin: .1rem 0; padding: .3rem .6rem; border: 0; border-radius: 6px; background: none;
    color: inherit; text-decoration: none; cursor: pointer; }}
  #stage-menu button:hover, #stage-menu a:hover {{ background: #8882; }}
  #stage-menu .head {{ font-size: .7rem; opacity: .6; padding: .2rem .6rem;
    text-transform: uppercase; letter-spacing: .06em; }}
  .node-stages {{ display: inline-flex; align-items: center; gap: .15rem; }}
  .node-stages .label {{ font-size: .78rem; opacity: .7; margin-left: .5rem; }}
  td.info {{ white-space: normal; min-width: 15rem; max-width: 24rem; font-size: .8rem; }}
  .icon {{ width: 1.35rem; height: 1.35rem; vertical-align: -.4em; fill: currentColor; opacity: .6; }}
  a:hover .icon {{ opacity: 1; }}
  td.deps code, .waits code {{ font-size: .78rem; padding: .1rem .4rem; border-radius: 4px; background: #8881; opacity: .55; }}
  td.deps code.unmet, .waits code.unmet {{ opacity: 1; background: #f9731633; font-weight: 600; }}
  td.deps a, .waits a {{ color: inherit; text-decoration: none; }}
  td.deps a:hover code, .waits a:hover code {{ opacity: 1; outline: 1px solid var(--line); }}
  .tree, .tree ul {{ list-style: none; margin: 0; padding: 0; font-size: .9rem; }}
  .tree ul {{ margin-left: 1.2rem; padding-left: 1.2rem; border-left: 1px solid var(--line); }}
  .tree li {{ position: relative; }}
  .tree ul > li::before {{ content: ""; position: absolute; left: -1.2rem; top: 1.15rem; width: 1rem; border-top: 1px solid var(--line); }}
  .node {{ display: flex; align-items: center; gap: .55rem; flex-wrap: wrap; padding: .3rem 0; scroll-margin-top: 1.5rem; }}
  .node:hover {{ background: #8881; }}
  .node:target {{ background: #3b82f61f; }}
  .node .name {{ opacity: .85; }}
  .node .waits {{ font-size: .78rem; opacity: .7; }}
  td.actions {{ text-align: right; }}
  .tag {{ font-size: .75rem; padding: .15rem .6rem; border: 1px solid var(--line); border-radius: 999px; }}
  .tag.completed {{ color: #16a34a; border-color: #16a34a66; background: #16a34a1a; font-weight: 600; }}
  .local {{ font-weight: 600; }}
  .note {{ opacity: .7; font-size: .8rem; margin-top: 1.75rem; max-width: 52rem; }}
  .feedback {{ position: sticky; top: 0; z-index: 2; background: Canvas; padding: .75rem; border: 1px solid var(--line); }}
  .feedback[data-error="true"] {{ border-color: #dc2626; color: #b91c1c; }}
  .stage-reason {{ display: block; white-space: normal; font-size: .8rem; max-width: 24rem; }}
  button:disabled {{ cursor: not-allowed; opacity: .55; }}
  form {{ display: inline; }}
  button {{ font: inherit; font-size: .8rem; padding: .3rem .8rem; margin-left: .4rem; cursor: pointer; border-radius: 6px; }}
  .menu {{ display: inline-block; position: relative; }}
  .menu > summary {{ list-style: none; cursor: pointer; font-size: .8rem; margin-left: .4rem;
    padding: .3rem .7rem; border: 1px solid var(--line); border-radius: 6px; }}
  .menu > summary::-webkit-details-marker {{ display: none; }}
  .menu > div {{ position: absolute; right: 0; z-index: 3; padding: .35rem; text-align: left;
    background: Canvas; border: 1px solid var(--line); border-radius: 8px; white-space: nowrap;
    display: flex; flex-direction: column; box-shadow: 0 2px 10px #0003; }}
  .menu > div button {{ margin: .1rem 0; text-align: left; }}
</style>
<header>
  <h1>Code Winch scheduler</h1>
  <button type="button" data-action="sync">Sync main</button>
  <label>API token <input id="token" type="password" autocomplete="off"></label>
  <button type="button" id="sign-in">Sign in</button>
  <button type="button" id="sign-out" hidden>Sign out</button>
  <span id="auth-status">Not signed in</span>
  <button type="button" id="swap" hidden>Tree view</button>
</header>
<div class="feedback" id="feedback" hidden><span id="result" role="alert" aria-live="assertive"></span>
<a id="refresh" href="" hidden>Refresh task list</a></div>
<div class="sub">{summary}</div>
{message}
<div class="wrap" id="table-view">
<table>
  <thead><tr><th>Task</th><th>Title</th><th>Pull request</th><th class="stage">Refine</th><th class="stage">Implement</th><th class="stage">Audit</th><th>Actions</th><th>Depends on</th><th>Status</th><th>Source</th><th>Updated</th><th>Info</th></tr></thead>
  <tbody>
  {rows}
  </tbody>
</table>
</div>
<ul class="tree wrap" id="tree-view" hidden>
  {tree}
</ul>
<div id="stage-menu" role="menu" hidden></div>
<p class="note">Each stage column is one control, marked with the agent that runs it: click its icon
to launch that stage, which turns it green, and click it again to open the conversation it started.
Right-click it for the rest — submit the stage again, expire the submission so the next click
launches afresh, or open a conversation an earlier attempt left behind. The GitHub icon is grey
until Sync main finds a pull request from the task branch into main, then green and linked;
click the grey one to create that pull request, or right-click to create one, open GitHub's
new-pull-request page, or refresh just this task. Creating needs the panel's GitHub App to hold
pull-request write permission on this repository.
Sync main prepares every
available task branch; Prepare claims one named task against a free scheduler slot. Refine and
Implement run Codex on the task branch; merge refinement changes there before starting
implementation. Audit starts a Claude cloud session for an open implementation pull request into
that branch. Repeating a stage at the same revision — the branch's, or the pull request's for
Audit — returns the conversation it already opened, so submit again after the revision moves on.
Expire releases a local reservation. No expiry cancels a cloud task or removes a link.</p>
<script>
  const result = document.getElementById("result");
  const feedback = document.getElementById("feedback");
  const refresh = document.getElementById("refresh");
  const token = document.getElementById("token");
  const authStatus = document.getElementById("auth-status");
  const signInButton = document.getElementById("sign-in");
  const signOutButton = document.getElementById("sign-out");
  // The page is the mount point, so resolve calls against it however it is proxied.
  const base = location.pathname.endsWith("/") ? location.pathname : location.pathname + "/";
  refresh.href = location.pathname;
  function signedIn(value) {{
    authStatus.textContent = value ? "Signed in on this browser" : "Not signed in";
    signOutButton.hidden = !value;
  }}
  function showMessage(message, error = false) {{
    feedback.hidden = false;
    feedback.dataset.error = String(error);
    result.textContent = message;
    feedback.scrollIntoView({{block: "nearest"}});
  }}
  async function request(path, body, bearer) {{
    if (!bearer && !window.isSecureContext) bearer = token.value;
    const headers = {{"X-Panel-Request": "1"}};
    if (body !== undefined) headers["Content-Type"] = "application/json";
    if (bearer) headers.Authorization = "Bearer " + bearer;
    const response = await fetch(base + path, {{method: body === undefined ? "GET" : "POST",
      credentials: "same-origin", headers, body: body === undefined ? undefined : JSON.stringify(body)}});
    const text = await response.text();
    let data;
    try {{ data = JSON.parse(text); }} catch (_) {{ data = null; }}
    if (response.status === 401) {{
      signedIn(false);
      token.focus();
      throw new Error("Unauthorized (401). Enter a valid API token and sign in, then retry.");
    }}
    if (!response.ok) {{
      const error = new Error(data?.error || "Request failed (HTTP " + response.status + "). Check the proxy and panel connection.");
      error.data = data;
      throw error;
    }}
    if (!data) throw new Error("The server returned an unexpected response (HTTP " + response.status + ").");
    return data;
  }}
  async function signIn() {{
    if (!window.isSecureContext) throw new Error("Browser sign-in requires HTTPS or localhost.");
    await request("api/session", {{path: base}}, token.value);
    token.value = "";
    const session = await request("api/session");
    if (!session.authenticated) throw new Error("The session cookie was not accepted. Use HTTPS and allow cookies for this site.");
    signedIn(true);
  }}
  signInButton.onclick = async () => {{
    try {{ await signIn(); showMessage("Signed in. This browser will remember the session for seven days."); }}
    catch (error) {{ showMessage(error.message, true); }}
  }};
  signOutButton.onclick = async () => {{
    try {{ await request("api/session/logout", {{path: base}}); token.value = ""; signedIn(false); showMessage("Signed out."); }}
    catch (error) {{ showMessage(error.message, true); }}
  }};
  request("api/session").then(data => signedIn(data.authenticated)).catch(error => showMessage(error.message, true));
  async function act(endpoint, body) {{
    if (token.value && window.isSecureContext) await signIn();
    return await request(endpoint, body === undefined ? {{}} : body);
  }}
  // Sync, Prepare and Expire all change what the whole page says, so each ends in a reload.
  document.querySelectorAll("[data-action]").forEach(button => {{
    button.onclick = async () => {{
      const action = button.dataset.action;
      const endpoint = action === "sync" ? "api/sync" :
        "api/tasks/" + encodeURIComponent(button.dataset.task) + "/" + action;
      button.disabled = true;
      refresh.hidden = true;
      showMessage("Working…");
      try {{
        let data = await act(endpoint);
        // Sync and Prepare both answer with a job envelope, not a result, and are polled.
        if (action === "sync" || action === "prepare") {{
          while (data.status === "running") {{
            showMessage("Syncing main and preparing task branches…");
            await new Promise(resolve => setTimeout(resolve, 1000));
            data = await request("api/sync/" + encodeURIComponent(data.id));
          }}
          if (data.status === "failed") throw new Error(data.error);
        }}
        window.location.reload();
      }} catch (error) {{
        showMessage(error.message, true);
        if (action === "sync") refresh.hidden = false;
      }} finally {{ button.disabled = false; }}
    }};
  }});
  const stageMenu = document.getElementById("stage-menu");
  const stageIcon = (cell) => cell.querySelector(".stage-icon");
  function closeMenu() {{
    stageMenu.hidden = true;
    stageMenu.textContent = "";
  }}
  function linkItem(href, text) {{
    const link = document.createElement("a");
    link.href = href; link.textContent = text;
    link.target = "_blank"; link.rel = "noopener";
    link.onclick = () => closeMenu();
    return link;
  }}
  function actionItem(text, hint, refusal, run) {{
    const entry = document.createElement("button");
    entry.type = "button";
    entry.textContent = text;
    entry.title = refusal || hint || "";
    if (refusal) entry.disabled = true;
    else entry.onclick = () => {{ closeMenu(); run(); }};
    return entry;
  }}
  // A launched stage becomes a plain link, so the next click opens it without the script.
  function becomeOpen(cell, url) {{
    const icon = stageIcon(cell);
    icon.href = url;
    icon.target = "_blank"; icon.rel = "noopener";
    icon.removeAttribute("role");
    icon.dataset.state = "open";
    icon.title = "Open the " + cell.dataset.stage + " conversation. Right-click for " +
      cell.dataset.stage + " actions.";
    cell.dataset.expireReason = "";
    cell.dataset.expireHint = "Submit " + cell.dataset.stage + " again at the same revision.";
  }}
  // Expiring keeps the conversation, which moves to the menu; the icon returns to launching.
  function becomeLaunch(cell) {{
    const icon = stageIcon(cell);
    const earlier = cell.querySelector(".earlier");
    if (icon.href && earlier) earlier.append(linkItem(icon.href, String(earlier.children.length + 1)));
    // Clearing the property is not enough in a browser: href="" resolves to this page.
    icon.href = "";
    icon.removeAttribute("href");
    icon.setAttribute("role", "button");
    icon.tabIndex = 0;
    icon.dataset.state = "launch";
    icon.title = cell.dataset.reason || cell.dataset.launch;
    cell.dataset.expireReason = "No " + cell.dataset.stage + " submission is waiting to be expired.";
  }}
  async function launchStage(cell) {{
    const stage = cell.dataset.stage;
    if (cell.dataset.reason) {{ showMessage(cell.dataset.reason, true); return; }}
    const endpoint = "api/tasks/" + encodeURIComponent(cell.dataset.task) + "/" + stage;
    showMessage("Working…");
    try {{
      let data;
      try {{ data = await act(endpoint); }}
      catch (error) {{
        if (stage !== "audit" || !error.data?.pull_requests?.length) throw error;
        const selected = window.prompt("Choose implementation PR number: " +
          error.data.pull_requests.map(p => p.number + ": " + p.url).join("\n"));
        if (!selected) throw error;
        data = await act(endpoint, {{pr_number: Number(selected)}});
      }}
      becomeOpen(cell, data.task_url);
      showMessage((data.reused ? "Already submitted: " : "Submitted: ") + stage + " — ");
      result.appendChild(linkItem(data.task_url, data.task_url));
    }} catch (error) {{ showMessage(error.message, true); }}
  }}
  async function expireStage(cell) {{
    const stage = cell.dataset.stage;
    showMessage("Working…");
    try {{
      await act("api/tasks/" + encodeURIComponent(cell.dataset.task) + "/" + stage + "/expire");
      becomeLaunch(cell);
      showMessage("Expired " + stage + " for " + cell.dataset.task +
        ". The next click submits it again; the conversation stays on the right-click menu.");
    }} catch (error) {{ showMessage(error.message, true); }}
  }}
  function openMenu(cell, x, y) {{
    const icon = stageIcon(cell);
    const stage = cell.dataset.stage;
    stageMenu.textContent = "";
    const heading = document.createElement("div");
    heading.className = "head";
    heading.textContent = stage + " · " + cell.dataset.task;
    stageMenu.append(heading);
    if (icon.href) stageMenu.append(linkItem(icon.href, "Open conversation"));
    stageMenu.append(actionItem("Submit " + stage, cell.dataset.launch, cell.dataset.reason,
      () => launchStage(cell)));
    stageMenu.append(actionItem(
      "Expire " + stage + (icon.dataset.state === "unknown" ? " (outcome unknown)" : ""),
      cell.dataset.expireHint, cell.dataset.expireReason, () => expireStage(cell)));
    const earlier = cell.querySelector(".earlier");
    Array.from(earlier ? earlier.children : []).forEach(link => {{
      stageMenu.append(linkItem(link.href, "Earlier conversation " + link.textContent));
    }});
    placeMenu(x, y);
  }}
  function placeMenu(x, y) {{
    stageMenu.hidden = false;
    stageMenu.style.left = (x + window.scrollX) + "px";
    stageMenu.style.top = (y + window.scrollY) + "px";
  }}
  function openPullRequestMenu(cell, x, y) {{
    stageMenu.textContent = "";
    const heading = document.createElement("div");
    heading.className = "head";
    heading.textContent = "pull request · " + cell.dataset.task;
    stageMenu.append(heading);
    if (cell.dataset.found) stageMenu.append(linkItem(cell.dataset.found, "Open pull request"));
    stageMenu.append(actionItem("Create pull request",
      "Open a new pull request from this task's branch into main, even if one exists. " +
        "GitHub refuses a second open one for the same branches.",
      cell.dataset.createReason, () => createPullRequest(cell, true)));
    stageMenu.append(cell.dataset.compare
      ? linkItem(cell.dataset.compare, "Open GitHub's new pull request page")
      : actionItem("Open GitHub's new pull request page", "", cell.dataset.compareReason, () => {{}}));
    stageMenu.append(actionItem("Refresh pull request",
      "Look for this task's pull request on GitHub now, without a full sync.",
      cell.dataset.refreshReason, () => refreshPullRequest(cell)));
    placeMenu(x, y);
  }}
  // Mirrors the server's rendering of the cell, so the answer shows without a reload.
  function showPullRequest(cell, url, state) {{
    const icon = cell.querySelector(".pr-icon");
    cell.dataset.found = url || "";
    if (url) {{
      icon.href = url; icon.target = "_blank"; icon.rel = "noopener";
      icon.removeAttribute("role");
      icon.dataset.state = "found";
      icon.title = "Open pull request #" + url.split("/").pop() +
        (state ? " (" + state.toLowerCase() + ")" : "");
    }} else {{
      icon.removeAttribute("href");
      icon.setAttribute("role", "button");
      icon.dataset.state = cell.dataset.createReason ? "blocked" : "none";
      icon.title = cell.dataset.createReason || "No pull request yet. Click to create one.";
    }}
  }}
  async function createPullRequest(cell, force = false) {{
    const task = cell.dataset.task;
    if (cell.dataset.createReason) {{ showMessage(cell.dataset.createReason, true); return; }}
    showMessage("Creating the pull request for " + task + "…");
    try {{
      const data = await act("api/tasks/" + encodeURIComponent(task) + "/create-pull-request",
        force ? {{force: true}} : undefined);
      showPullRequest(cell, data.pull_request, data.pull_request_state);
      // One was already open, so this click is the green icon's click: take the operator to it.
      if (!data.created) window.open(data.pull_request, "_blank", "noopener");
      showMessage((data.created ? "Opened pull request: " : "Already open, opening it: ") +
        data.pull_request);
    }} catch (error) {{ showMessage(error.message, true); }}
  }}
  async function refreshPullRequest(cell) {{
    const task = cell.dataset.task;
    showMessage("Looking for the pull request of " + task + "…");
    try {{
      const data = await act("api/tasks/" + encodeURIComponent(task) + "/pull-request");
      showPullRequest(cell, data.pull_request, data.pull_request_state);
      showMessage(data.pull_request ? "Pull request for " + task + ": " + data.pull_request
                                    : "No pull request from task/" + task + " into main yet.");
    }} catch (error) {{ showMessage(error.message, true); }}
  }}
  document.querySelectorAll("[data-pr-cell]").forEach(cell => {{
    const icon = cell.querySelector(".pr-icon");
    // A found pull request is a plain link; without one the click creates it.
    const create = (event) => {{
      if (icon.href) return;
      event.preventDefault();
      createPullRequest(cell);
    }};
    icon.onclick = create;
    icon.onkeydown = (event) => {{
      if (event.key === "Enter" || event.key === " ") create(event);
    }};
    cell.oncontextmenu = (event) => {{
      event.preventDefault();
      openPullRequestMenu(cell, event.clientX, event.clientY);
    }};
  }});
  document.querySelectorAll("[data-stage-cell]").forEach(cell => {{
    const icon = stageIcon(cell);
    const launch = (event) => {{
      if (icon.href) return;
      event.preventDefault();
      launchStage(cell);
    }};
    icon.onclick = launch;
    icon.onkeydown = (event) => {{
      if (event.key === "Enter" || event.key === " ") launch(event);
    }};
    icon.oncontextmenu = (event) => {{
      event.preventDefault();
      openMenu(cell, event.clientX, event.clientY);
    }};
  }});
  document.addEventListener("click", (event) => {{
    if (!stageMenu.hidden && !stageMenu.contains(event.target)) closeMenu();
  }});
  document.addEventListener("keydown", (event) => {{
    if (event.key === "Escape") closeMenu();
  }});
  const swap = document.getElementById("swap");
  const table = document.getElementById("table-view");
  const tree = document.getElementById("tree-view");
  const show = (view) => {{
    table.hidden = view === "tree";
    tree.hidden = view !== "tree";
    swap.textContent = view === "tree" ? "Table view" : "Tree view";
    localStorage.setItem("code-winch-view", view);
  }};
  show(localStorage.getItem("code-winch-view") === "tree" ? "tree" : "table");
  swap.hidden = false;
  swap.onclick = () => show(tree.hidden ? "tree" : "table");
</script>
"""

NODE = """<li>
  <div class="node" id="tree-{id}">
    <code>{id}</code><span class="name">{title}</span>
    <span class="tag {status}">{status}</span>{waits}
    <span class="node-stages"><span class="label">Refine</span>{refine}<span
      class="label">Implement</span>{implement}<span class="label">Audit</span>{audit}</span>{actions}
    <span class="stage-reason">{info}</span>
  </div>
  {children}
</li>"""

ROW = """<tr id="{id}">
  <td><code>{id}</code></td>
  <td class="title">{title}</td>
  <td class="pr">{pull_request}</td>
  <td class="stage">{refine}</td>
  <td class="stage">{implement}</td>
  <td class="stage">{audit}</td>
  <td class="actions">{actions}</td>
  <td class="deps">{deps}</td>
  <td><span class="tag {status}">{status}</span></td>
  <td>{source}</td>
  <td>{updated}</td>
  <td class="info">{info}</td>
</tr>"""


# The OpenAI mark, inlined so the panel keeps working without any outbound request.
CODEX_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Codex conversation">'
    '<path d="M22.282 9.821a5.985 5.985 0 0 0-.516-4.911 6.046 6.046 0 0 0-6.51-2.9A6.065 6.065 0'
    " 0 0 4.981 4.182a5.985 5.985 0 0 0-3.998 2.9 6.046 6.046 0 0 0 .743 7.097 5.98 5.98 0 0 0"
    " .51 4.911 6.051 6.051 0 0 0 6.515 2.9A5.985 5.985 0 0 0 13.26 24a6.056 6.056 0 0 0"
    " 5.772-4.206 5.99 5.99 0 0 0 3.998-2.9 6.056 6.056 0 0 0-.748-7.073zm-9.022 12.608a4.476"
    " 4.476 0 0 1-2.876-1.04l.142-.08 4.778-2.759a.795.795 0 0 0 .393-.681v-6.737l2.02"
    " 1.169a.071.071 0 0 1 .038.052v5.583a4.504 4.504 0 0 1-4.495 4.494zm-9.66-4.126a4.471 4.471"
    " 0 0 1-.535-3.014l.142.086 4.783 2.758a.771.771 0 0 0 .78 0l5.843-3.368v2.332a.08.08 0 0"
    " 1-.033.062L9.74 19.95a4.499 4.499 0 0 1-6.14-1.647zM2.34 7.896a4.485 4.485 0 0 1"
    " 2.366-1.973v5.677a.766.766 0 0 0 .388.677l5.815 3.354-2.02 1.169a.076.076 0 0"
    " 1-.071 0l-4.83-2.787A4.504 4.504 0 0 1 2.34 7.872zm16.597 3.856L13.104 8.364 15.119"
    " 7.2a.076.076 0 0 1 .071 0l4.83 2.791a4.494 4.494 0 0 1-.676 8.105v-5.678a.79.79 0 0"
    " 0-.407-.666zm2.01-3.023l-.141-.085-4.774-2.782a.776.776 0 0 0-.785 0L9.41"
    " 9.23V6.897a.066.066 0 0 1 .028-.061l4.83-2.787a4.499 4.499 0 0 1 6.68 4.66zM8.307"
    " 12.863l-2.02-1.164a.08.08 0 0 1-.038-.057V6.074a4.499 4.499 0 0 1 7.376-3.454l-.142"
    " .08-4.778 2.76a.795.795 0 0 0-.393.68zm1.097-2.366l2.602-1.5 2.607 1.5v2.999l-2.597"
    ' 1.5-2.607-1.5z"/></svg>'
)

# The GitHub mark, inlined for the same reason.
GITHUB_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Pull request">'
    '<path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577'
    " 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633"
    " 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305"
    " 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93"
    " 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267"
    " 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24"
    " 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81"
    " 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0"
    '-6.627-5.373-12-12-12"/></svg>'
)

# An approximation of the Claude mark, a burst of uneven rays, drawn here rather than copied
# from the brand asset. Replace the body with the official path if you have it.
CLAUDE_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Claude session">'
    + "".join(
        f'<path transform="rotate({30 * index} 12 12)" d="M11.1 12 11.5 {12 - length}h1l.4 {length}z"/>'
        for index, length in enumerate((10, 7.5, 9, 8, 10.5, 7, 9.5, 8, 10, 7.5, 9, 8.5))
    )
    + "</svg>"
)

# A circled cross for releasing a task's scheduler slot.
EXPIRE_ICON = (
    '<svg class="icon" viewBox="0 0 24 24" role="img" aria-label="Expire">'
    '<circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="2"/>'
    '<path d="m8.5 8.5 7 7m0-7-7 7" fill="none" stroke="currentColor" stroke-width="2" '
    'stroke-linecap="round"/></svg>'
)

STAGES = ("refine", "implement", "audit")
# Who runs each stage, named where an uncertain submission must be checked by hand.
SERVICES = {"refine": "Codex", "implement": "Codex", "audit": "Claude"}


def button(action: str, task_id: str, label: str, hint: str = "", css: str = "") -> str:
    attributes = f' title="{html.escape(hint)}"' if hint else ""
    if css:
        attributes += f' class="{css}"'
    return (f'<button type="button" data-action="{html.escape(action)}" '
            f'data-task="{html.escape(task_id)}"{attributes}>{label}</button>')


def preparable(entry: dict[str, Any], runnable: set[str]) -> bool:
    """Whether the panel can admit this task now, by request rather than by scheduling pass.

    An expired reservation lands here too: the overlay stops applying, so the entry reads at
    the tracker's status while keeping the branch it already prepared.
    """
    if entry["status"] == "completed" or (entry["local"] and entry["prepared"]):
        return False
    return entry["status"] == "in_progress" or entry["id"] in runnable


def stage_reason(entry: dict[str, Any], runnable: set[str]) -> str:
    if entry["local"] and entry["prepared"] and entry["status"] == "in_progress":
        return ""
    if entry["status"] == "completed":
        reason = "Task completed."
    elif entry.get("prepare_error"):
        reason = entry["prepare_error"]
    elif preparable(entry, runnable):
        reason = "Prepare this task before choosing a stage; preparation needs a free slot."
    else:
        reason = "Waiting for dependencies or a blocked task to be released."
    return reason


def attempts_for(entry: dict[str, Any], stage: str) -> list[dict[str, Any]]:
    """This stage's submissions, oldest first."""
    return entry["stages"].get(stage, [])


def actions_for(entry: dict[str, Any], runnable: set[str]) -> str:
    """Whole-task actions. Launching a stage belongs to that stage's own cell."""
    actions = ""
    if preparable(entry, runnable):
        actions += button("prepare", entry["id"], "Prepare",
                          hint="Refresh main, claim a scheduler slot, and prepare this branch.")
    if entry["local"] and entry["status"] != "completed":
        actions += button("expire", entry["id"], EXPIRE_ICON,
                          hint="Release this task's scheduler slot. Stage records and the "
                               "conversations they opened are kept.", css="icon-button")
    return actions or "—"


def standing(entry: dict[str, Any], stage: str) -> tuple[dict[str, Any] | None, list[dict[str, Any]]]:
    """This stage's live attempt, and the earlier ones that still carry a link, newest first.

    The page cannot fetch the branch revision, so the newest unexpired attempt is taken as
    the live one; whether it actually covers the revision the branch is on now is decided by
    the panel on submission. An expired attempt keeps its link, because retiring a stage must
    not lose the conversation it opened — the right-click menu is where those stay reachable.
    """
    attempts = attempts_for(entry, stage)
    live = attempts[-1] if attempts and attempts[-1]["status"] != "expired" else None
    earlier = [attempt for attempt in attempts
               if attempt is not live and link_target(attempt.get("task_url"))]
    return live, earlier[::-1]


def expiry_of(stage: str, live: dict[str, Any] | None) -> tuple[str, str]:
    """Why this stage's submission cannot be expired, or what expiring it will do."""
    if live is None:
        return f"No {stage} submission is waiting to be expired.", ""
    if live["status"] == "submitting":
        # Not an override but a verdict on an unknown, and the risk runs the other way.
        return "", (f"{SERVICES[stage]} may have accepted this before the panel lost the reply. "
                    f"Check {SERVICES[stage]} first: expiring can submit it twice.")
    revision = "pull request head" if stage == "audit" else "branch revision"
    return "", f"Submit {stage} again at the same {revision}."


def stage_cell(entry: dict[str, Any], stage: str, runnable: set[str]) -> str:
    """One control for the stage: launch it, then open what it started.

    The icon is the mark of the agent that runs the stage, so a column is read by the shape in
    it rather than by a label. It launches while there is nothing to open and is a plain link
    once there is, which keeps the second click — and middle-click, and copy-link — native.
    Everything else about the stage lives behind the right-click menu the script builds from
    the data here.
    """
    live, earlier = standing(entry, stage)
    # Records written before /remote/ was the task view still hold a retired path.
    url = canonical_task_url(link_target(live.get("task_url"))) if live else ""
    reason = stage_reason(entry, runnable)
    expire_reason, expire_hint = expiry_of(stage, live)
    launch = f"Launch {stage} on {SERVICES[stage]}. Right-click for {stage} actions."
    if url:
        state, label = "open", f"open the {stage} conversation"
        title = f"Open the {stage} conversation. Right-click for {stage} actions."
        anchor = f'href="{html.escape(url)}" target="_blank" rel="noopener"'
    elif live is not None:
        state, label = "unknown", f"{stage} outcome unknown"
        title = (f"{SERVICES[stage]} may have accepted this {stage} before the panel lost the "
                 f"reply. Check {SERVICES[stage]}, then expire it from the right-click menu.")
        anchor = 'role="button" tabindex="0"'
    else:
        state, label = "launch", f"launch {stage}"
        title, anchor = reason or launch, 'role="button" tabindex="0"'
    history = "".join(
        f'<a href="{html.escape(canonical_task_url(link_target(attempt["task_url"])))}" '
        f'target="_blank" rel="noopener">{index}</a>'
        for index, attempt in enumerate(earlier, 1))
    return (
        f'<span class="stage-cell" data-stage-cell data-task="{html.escape(entry["id"])}" '
        f'data-stage="{stage}" data-service="{SERVICES[stage]}" '
        f'data-reason="{html.escape(reason)}" data-launch="{html.escape(launch)}" '
        f'data-expire-reason="{html.escape(expire_reason)}" '
        f'data-expire-hint="{html.escape(expire_hint)}">'
        f'<a class="stage-icon" data-state="{state}" {anchor} title="{html.escape(title)}" '
        f'aria-label="{html.escape(label)}">{CLAUDE_ICON if stage == "audit" else CODEX_ICON}</a>'
        f'<span class="earlier" hidden>{history}</span></span>'
    )


def source_of(entry: dict[str, Any]) -> str:
    return '<span class="local">lease</span>' if entry["local"] else "tracker"


def link_target(value: Any) -> str:
    """Return the value as a link target, or an empty string when it is not one.

    The scheduler stores the raw dispatch output under task_url when no URL matched, so the value
    only becomes a link once it parses as http(s) — which also keeps a javascript: href out of the
    page if the state file is ever hand-edited.
    """
    url = str(value or "").strip()
    parsed = urlparse(url)
    return url if parsed.scheme in ("http", "https") and parsed.netloc else ""


def pull_request_blocker(entry: dict[str, Any]) -> str:
    """Why a pull request cannot be opened for this task now, or an empty string."""
    if entry["status"] == "completed":
        return "Task completed."
    if not entry["prepared"]:
        return "Prepare this task first; its branch must exist before a pull request can open."
    return ""


def compare_url(entry: dict[str, Any], repository_url: str) -> tuple[str, str]:
    """GitHub's new-pull-request page for the task branch into main, or why there is none."""
    blocker = pull_request_blocker(entry)
    base = link_target(repository_url).rstrip("/").removesuffix(".git")
    if blocker:
        return "", blocker
    if not base:
        return "", "Set GITHUB_URL on the control panel to open this page."
    return f"{base}/compare/main...{quote(task_branch(entry['id']), safe='/')}?expand=1", ""


def pull_request_cell(entry: dict[str, Any], repository_url: str = "") -> str:
    """The pull request into main: green and linked once found, otherwise click to create it.

    The icon is always drawn. The right-click menu can create or open GitHub's page for a new
    pull request even when one exists, since the first may have been closed or not be the one
    the operator wants.
    """
    found = link_target(entry["pull_request"])
    compare, compare_reason = compare_url(entry, repository_url)
    blocker = pull_request_blocker(entry)
    if found:
        number = urlparse(found).path.rstrip("/").rpartition("/")[2]
        state = str(entry.get("pull_request_state") or "").lower()
        hint = f"Open pull request #{number}" if number.isdigit() else "Open the pull request"
        hint += f" ({state})" if state else ""
        mark = "found"
        anchor = f'href="{html.escape(found)}" target="_blank" rel="noopener"'
    elif blocker:
        hint, mark, anchor = blocker, "blocked", ""
    else:
        hint, mark = "No pull request yet. Click to create one.", "none"
        anchor = 'role="button" tabindex="0"'
    return (
        f'<span class="pr-cell" data-pr-cell data-task="{html.escape(entry["id"])}" '
        f'data-found="{html.escape(found)}" data-compare="{html.escape(compare)}" '
        f'data-compare-reason="{html.escape(compare_reason)}" '
        f'data-create-reason="{html.escape(blocker)}" '
        f'data-refresh-reason="{"" if entry["prepared"] else "Prepare this task first; there is no branch to look on."}">'
        f'<a class="pr-icon" data-state="{mark}" {anchor} '
        f'title="{html.escape(hint + (". Right-click for more." if mark != "blocked" else ""))}">'
        f'{GITHUB_ICON}</a></span>'
    )


def chips(dependencies: list[str], done: set[str], anchor: str) -> str:
    return " ".join(
        f'<a href="#{anchor}{html.escape(dependency)}" title="jump to {html.escape(dependency)}">'
        f'<code class="{"" if dependency in done else "unmet"}">{html.escape(dependency)}</code>'
        "</a>"
        for dependency in dependencies
    )


def branches(nodes: list[dict[str, Any]], runnable: set[str], done: set[str]) -> str:
    items = []
    for node in nodes:
        entry = node["entry"]
        waits = chips(node["waits_on"], done, "tree-")
        items.append(
            NODE.format(
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                status=html.escape(entry["status"]),
                waits=f'<span class="waits">also waits on {waits}</span>' if waits else "",
                actions=actions_for(entry, runnable),
                **{stage: stage_cell(entry, stage, runnable) for stage in STAGES},
                info=html.escape(stage_reason(entry, runnable)),
                children=(
                    f'<ul>{branches(node["children"], runnable, done)}</ul>'
                    if node["children"]
                    else ""
                ),
            )
        )
    return "\n".join(items)


def render(tracker: dict[str, Any], state: dict[str, Any], message: str, busy: bool,
           repository_url: str = "") -> str:
    entries = rows(tracker, state)
    runnable = available_ids(entries)
    done = {entry["id"] for entry in entries if entry["status"] == "completed"}
    counts: dict[str, int] = {name: 0 for name in STATUS_ORDER}
    for entry in entries:
        counts[entry["status"]] = counts.get(entry["status"], 0) + 1
    summary = " · ".join(f"{counts[name]} {name.replace('_', ' ')}" for name in STATUS_ORDER)
    summary += f" · {len(runnable)} available"
    if busy:
        summary += " · scheduler running"

    body = []
    for entry in entries:
        deps = chips(entry["depends_on"], done, "")
        body.append(
            ROW.format(
                id=html.escape(entry["id"]),
                title=html.escape(entry["title"]),
                deps=deps or "—",
                status=html.escape(entry["status"]),
                source=source_of(entry),
                **{stage: stage_cell(entry, stage, runnable) for stage in STAGES},
                info=html.escape(stage_reason(entry, runnable)) or "—",
                pull_request=pull_request_cell(entry, repository_url),
                updated=html.escape((entry["updated_at"] or "—")[:16].replace("T", " ")),
                actions=actions_for(entry, runnable),
            )
        )

    banner = f'<div class="msg">{html.escape(message)}</div>' if message else ""
    return PAGE.format(
        summary=html.escape(summary),
        message=banner,
        rows="\n  ".join(body),
        tree=branches(forest(entries), runnable, done),
    )
