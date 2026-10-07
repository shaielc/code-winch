// Exercise the rendered browser script with HTTP/DOM doubles; no external packages.
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const vm = require('node:vm');
const page = execFileSync('python3', ['-c',
  'from control_panel.ui import render; print(render({"tasks": []}, {"tasks": {}}, "", False))'], {encoding: 'utf8'});
const script = page.match(/<script>([\s\S]*?)<\/script>/)[1];
function element() {
  const node = {dataset: {}, style: {}, attributes: {}, value: '', hidden: false, children: [],
    append(...items) { this.children.push(...items); }, appendChild(item) { this.children.push(item); },
    setAttribute(name, value) { this.attributes[name] = value; },
    removeAttribute(name) { delete this.attributes[name]; },
    contains(item) { return this.children.includes(item); },
    querySelector() {return null;}, select() {}, focus() {this.focused = true;},
    scrollIntoView() {this.scrolled = true;}};
  let text = '';
  // Emptying textContent clears an element in a browser, which is how the menu is rebuilt.
  Object.defineProperty(node, 'textContent', {
    get: () => text,
    set: (value) => { text = String(value); if (text === '') node.children = []; },
  });
  return node;
}
// The server renders these attributes; the stage spec mirrors one row of the table.
function stageCell(spec) {
  const cell = element(), icon = element(), earlier = element();
  cell.dataset = {stageCell: '', task: 'P0-001', stage: spec.stage, service: spec.service,
    reason: spec.reason || '',
    launch: 'Launch ' + spec.stage + ' on ' + spec.service + '. Right-click for actions.',
    expireReason: spec.href ? '' : 'No ' + spec.stage + ' submission is waiting to be expired.',
    expireHint: 'Submit ' + spec.stage + ' again at the same branch revision.'};
  icon.dataset.state = spec.href ? 'open' : 'launch';
  icon.href = spec.href || '';
  cell.children = [icon, earlier];
  cell.querySelector = selector => selector === '.earlier' ? earlier : icon;
  return cell;
}
const event = () => ({preventDefault() {this.prevented = true;}, clientX: 12, clientY: 34});
function browser(pathname, respond, secure = true) {
  const elements = new Map(), requests = [], listeners = {};
  const buttons = ['sync', 'prepare', 'expire'].map(
    action => Object.assign(element(), {dataset: {action, task: 'P0-001'}}));
  const stageCells = [
    {stage: 'refine', service: 'Codex'},
    {stage: 'implement', service: 'Codex', href: 'https://chatgpt.com/remote/task_e_impl'},
    {stage: 'audit', service: 'Claude', reason: 'Prepare this task before choosing a stage.'},
  ].map(stageCell);
  const get = id => {if (!elements.has(id)) elements.set(id, element()); return elements.get(id);};
  const location = {pathname, reload() {this.reloaded = true;}};
  const context = {location, window: {location, isSecureContext: secure, scrollX: 0, scrollY: 0},
    document: {getElementById: get, createElement: element,
      addEventListener(type, handler) { (listeners[type] = listeners[type] || []).push(handler); },
      querySelectorAll: selector => selector === '[data-action]' ? buttons
        : selector === '[data-stage-cell]' ? stageCells : []},
    localStorage: {getItem() {return 'table';}, setItem(key) {assert.equal(key, 'code-winch-view');}},
    setTimeout: fn => queueMicrotask(fn),
    fetch: async (url, options) => {
      requests.push({url, options});
      const reply = respond(url, options);
      return {ok: reply.status >= 200 && reply.status < 300, status: reply.status,
        text: async () => reply.text ?? JSON.stringify(reply.data)};
    }};
  vm.runInNewContext(script, context);
  const fire = (type, detail) => (listeners[type] || []).forEach(handler => handler(detail));
  return {get, buttons, requests, stageCells, location, fire};
}
const flush = () => new Promise(resolve => setImmediate(resolve));
(async () => {
  for (const prefix of ['/', '/panel', '/tools/panel/']) {
    const base = prefix.endsWith('/') ? prefix : prefix + '/';
    let remembered = false;
    const client = browser(prefix, (url, options) => {
      assert.ok(url.startsWith(base));
      const route = url.slice(base.length);
      if (route === 'api/session') {
        if (options.method === 'POST') {
          assert.equal(options.headers.Authorization, 'Bearer valid-token');
          assert.equal(JSON.parse(options.body).path, base);
          remembered = true;
        }
        return {status: 200, data: {authenticated: remembered}};
      }
      assert.equal(options.headers.Authorization, undefined);
      assert.equal(options.credentials, 'same-origin');
      assert.equal(options.headers['X-Panel-Request'], '1');
      if (route === 'api/session/logout') {remembered = false; return {status: 200, data: {authenticated: false}};}
      // Prepare is polled like a sync; expiring a stage answers directly.
      if (route === 'api/sync' || route === 'api/tasks/P0-001/prepare') {
        return {status: 202, data: {id: 'job1', status: 'running'}};
      }
      if (route === 'api/sync/job1') return {status: 200, data: {id: 'job1', status: 'succeeded'}};
      if (route === 'api/tasks/P0-001/refine/expire') {
        return {status: 200, data: {expired: 'P0-001', stage: 'refine'}};
      }
      return {status: 200, data: {task_url: 'https://chatgpt.com/remote/task_e_' + route.split('/').pop()}};
    });
    await flush();
    client.get('token').value = 'valid-token';
    await client.get('sign-in').onclick();
    assert.equal(client.get('token').value, '');
    assert.equal(client.get('auth-status').textContent, 'Signed in on this browser');
    // Sync, Prepare and the lease Expire each reload rather than patching the page.
    for (const button of client.buttons) await button.onclick();
    assert.ok(client.location.reloaded);
    assert.equal(client.get('feedback').dataset.error, 'false');
    const paths = client.requests.map(request => request.url.slice(base.length));
    assert.ok(paths.includes('api/tasks/P0-001/prepare'));
    assert.ok(paths.includes('api/tasks/P0-001/expire'));
    assert.equal(paths.filter(path => path === 'api/sync/job1').length, 2);

    const [refine, implement, audit] = client.stageCells;
    const icon = cell => cell.querySelector('.stage-icon');
    // The first click on a dim icon launches the stage and turns it into a link.
    let before = client.requests.length;
    icon(refine).onclick(event());
    await flush();
    assert.equal(client.requests.length, before + 1);
    assert.equal(client.requests.at(-1).url.slice(base.length), 'api/tasks/P0-001/refine');
    assert.equal(icon(refine).dataset.state, 'open');
    assert.equal(icon(refine).href, 'https://chatgpt.com/remote/task_e_refine');
    assert.equal(refine.dataset.expireReason, '');
    // The second click is the link's own: the script stays out of the way.
    icon(refine).onclick(event());
    await flush();
    assert.equal(client.requests.length, before + 1);
    // A stage rendered already-submitted behaves the same way on its first click.
    icon(implement).onclick(event());
    await flush();
    assert.equal(client.requests.length, before + 1);

    // Right-click offers the stage's actions, with the conversation first.
    const menu = client.get('stage-menu');
    icon(refine).oncontextmenu(event());
    assert.equal(menu.hidden, false);
    assert.equal(menu.style.left, '12px');
    const labels = menu.children.map(child => child.textContent);
    assert.deepEqual(labels, ['refine · P0-001', 'Open conversation', 'Submit refine', 'Expire refine']);
    const expire = menu.children.find(child => child.textContent === 'Expire refine');
    assert.equal(expire.disabled, undefined);
    before = client.requests.length;
    expire.onclick();
    await flush();
    assert.equal(menu.hidden, true);
    assert.equal(client.requests.at(-1).url.slice(base.length), 'api/tasks/P0-001/refine/expire');
    // Expiring returns the icon to launching without losing what it opened.
    assert.equal(icon(refine).dataset.state, 'launch');
    assert.equal(icon(refine).href, '');
    assert.equal(refine.querySelector('.earlier').children.length, 1);
    assert.equal(refine.querySelector('.earlier').children[0].href,
                 'https://chatgpt.com/remote/task_e_refine');
    icon(refine).oncontextmenu(event());
    assert.deepEqual(menu.children.map(child => child.textContent),
                     ['refine · P0-001', 'Submit refine', 'Expire refine', 'Earlier conversation 1']);
    assert.equal(menu.children.find(child => child.textContent === 'Expire refine').disabled, true);
    // A click anywhere else closes it, and so does Escape.
    client.fire('click', {target: element()});
    assert.equal(menu.hidden, true);
    icon(refine).oncontextmenu(event());
    assert.equal(menu.hidden, false);
    client.fire('keydown', {key: 'Escape'});
    assert.equal(menu.hidden, true);

    // An unprepared stage reports why instead of submitting anything.
    before = client.requests.length;
    icon(audit).onclick(event());
    await flush();
    assert.equal(client.requests.length, before);
    assert.equal(client.get('feedback').dataset.error, 'true');
    assert.ok(client.get('result').textContent.includes('Prepare this task'));
    await client.get('sign-out').onclick();
    assert.equal(client.get('auth-status').textContent, 'Not signed in');
  }
  for (const failure of [
    {status: 401, text: '<html>Unauthorized</html>'},
    {status: 502, text: '<html>Bad Gateway</html>'},
  ]) {
    const client = browser('/panel/', url => url.endsWith('api/session')
      ? {status: 200, data: {authenticated: false}} : failure);
    await flush();
    const icon = client.stageCells[0].querySelector('.stage-icon');
    icon.onclick(event());
    await flush();
    assert.equal(client.get('feedback').hidden, false);
    assert.equal(client.get('feedback').dataset.error, 'true');
    assert.ok(client.get('result').textContent.includes(String(failure.status)));
    if (failure.status === 401) assert.ok(client.get('token').focused);
    assert.equal(icon.dataset.state, 'launch');
  }
  const failedSync = browser('/panel', url => {
    if (url.endsWith('api/session')) return {status: 200, data: {authenticated: true}};
    if (url.endsWith('api/sync')) return {status: 202, data: {id: 'job', status: 'running'}};
    return {status: 200, data: {id: 'job', status: 'failed', error: 'Main refreshed; git push failed'}};
  });
  await flush();
  await failedSync.buttons[0].onclick();
  assert.equal(failedSync.location.reloaded, undefined);
  assert.equal(failedSync.get('result').textContent, 'Main refreshed; git push failed');
  assert.equal(failedSync.get('refresh').hidden, false);
  const temporary = browser('/panel', (url, options) => {
    if (!url.endsWith('api/session')) assert.equal(options.headers.Authorization, 'Bearer temporary-token');
    return {status: 200, data: {authenticated: false, task_url: 'https://chatgpt.com/remote/task_e_x'}};
  }, false);
  await flush();
  temporary.get('token').value = 'temporary-token';
  temporary.stageCells[0].querySelector('.stage-icon').onclick(event());
  await flush();
  assert.equal(temporary.get('feedback').dataset.error, 'false');
  console.log('UI: proxy paths, session login/logout, stage launch/open/expire, context menu, '
    + 'visible 401/502, and async sync passed.');
})().catch(error => {console.error(error); process.exitCode = 1;});
