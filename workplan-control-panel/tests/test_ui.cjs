// Exercise the rendered browser script with HTTP/DOM doubles; no external packages.
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const vm = require('node:vm');
const page = execFileSync('python3', ['-c',
  'from control_panel.ui import render; print(render({"tasks": []}, {"tasks": {}}, "", False))'], {encoding: 'utf8'});
const script = page.match(/<script>([\s\S]*?)<\/script>/)[1];
function element() {
  return {dataset: {}, value: '', hidden: false, textContent: '', children: [],
    append(...items) { this.children.push(...items); }, appendChild(item) { this.children.push(item); },
    querySelector() {return null;}, select() {}, focus() {this.focused = true;}, scrollIntoView() {this.scrolled = true;}};
}
function browser(pathname, respond, secure = true) {
  const elements = new Map(), requests = [];
  const buttons = ['sync', 'refine', 'implement', 'audit', 'expire'].map(action =>
    Object.assign(element(), {dataset: {action, task: 'P0-001'}}));
  const conversations = ['refine', 'implement', 'refine', 'implement'].map(stage => {
    const container = element(), link = element(), placeholder = element();
    container.dataset = {conversations: 'P0-001', conversationStage: stage};
    link.dataset.stage = stage;
    link.hidden = true;
    container.children = [link, placeholder];
    container.querySelector = selector => selector === '.no-conversation' ? placeholder : link;
    return container;
  });
  const get = id => {if (!elements.has(id)) elements.set(id, element()); return elements.get(id);};
  const location = {pathname, reload() {this.reloaded = true;}};
  const context = {location, window: {location, isSecureContext: secure},
    document: {getElementById: get, createElement: element,
      querySelectorAll: selector => selector === '[data-action]' ? buttons : conversations},
    localStorage: {getItem() {return 'table';}, setItem(key) {assert.equal(key, 'code-winch-view');}},
    navigator: {clipboard: {async writeText() {}}}, setTimeout: fn => queueMicrotask(fn),
    fetch: async (url, options) => {
      requests.push({url, options});
      const reply = respond(url, options);
      return {ok: reply.status >= 200 && reply.status < 300, status: reply.status,
        text: async () => reply.text ?? JSON.stringify(reply.data)};
    }};
  vm.runInNewContext(script, context);
  return {get, buttons, requests, conversations, location};
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
      if (route === 'api/sync') return {status: 202, data: {id: 'job1', status: 'running'}};
      if (route === 'api/sync/job1') return {status: 200, data: {id: 'job1', status: 'succeeded'}};
      return {status: 200, data: {task_url: 'https://chatgpt.com/codex/tasks/' + route.split('/').pop(), prompt: 'Audit'}};
    });
    await flush();
    client.get('token').value = 'valid-token';
    await client.get('sign-in').onclick();
    assert.equal(client.get('token').value, '');
    assert.equal(client.get('auth-status').textContent, 'Signed in on this browser');
    for (const button of client.buttons) await button.onclick();
    assert.ok(client.location.reloaded);
    for (const container of client.conversations) {
      const [link, placeholder] = container.children;
      assert.equal(link.href, 'https://chatgpt.com/codex/tasks/' + container.dataset.conversationStage);
      assert.equal(link.hidden, false);
      assert.equal(placeholder.hidden, true);
    }
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
    await client.buttons[1].onclick();
    assert.equal(client.get('feedback').hidden, false);
    assert.equal(client.get('feedback').dataset.error, 'true');
    assert.ok(client.get('result').textContent.includes(String(failure.status)));
    if (failure.status === 401) assert.ok(client.get('token').focused);
    assert.equal(client.buttons[1].disabled, false);
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
    return {status: 200, data: {authenticated: false, task_url: 'https://chatgpt.com/codex/tasks/refine'}};
  }, false);
  await flush();
  temporary.get('token').value = 'temporary-token';
  await temporary.buttons[1].onclick();
  assert.equal(temporary.get('feedback').dataset.error, 'false');
  console.log('UI: proxy paths, session login/logout, stage links, visible 401/502, and async sync passed.');
})().catch(error => {console.error(error); process.exitCode = 1;});
