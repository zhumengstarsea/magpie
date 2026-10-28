// Run with Node's test runner and Playwright on the module path; see README.md.
// #2679 (@minggang123 on X): the Library's MCP section has a switch that
// gives each agent one MCP server, magpie, in place of the library's
// servers. Off, it says what it would do; a click posts servers/hub on,
// says the agents are given magpie's one server, and the page then says so.
// Each agent that keeps each server's own entry is named under it with why
// (Claude Desktop reaches no server over HTTP; Codex has a "magpie" of its
// own). On, it counts what it serves (servers, tools, the agents given it)
// and shows the one address any other app connects at, with Connect…: a
// sheet listing every agent magpie found that takes MCP servers, each with
// where it stands (connected, why it keeps each server's own entry, or a
// Connect that gives it every server, posting servers/agents-all, after
// which it says connected), then that address and, for Cursor (when it
// isn't one of the agents), VS Code and any JSON app, what adds it. Cursor
// and VS Code get an "Add to" button in the app (it posts servers/hub-open
// with the app's name, never a link) and none in a browser. No agent is
// told to the user as a command to run. The address is the public one in a
// browser and the gateway in the app. A
// click again turns it off. In every language, both engines, a wide window
// in the app and a narrow one in a browser: no sideways scroll, and the
// switch beside its text.
// No backend: the API is faked here.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");
const { strings } = require("./strings.cjs");

const assets = path.resolve(__dirname, "../assets");
const HOME = "/Users/mg";
const fixture = (on, st = {}) => ({
  dir: `${HOME}/.magpie/library`, backups: `${HOME}/.magpie/backups`, home: HOME,
  mcpHub: on, hubPath: "/mcp/magpie/hub/",
  agents: [
    { id: "claude", name: "Claude Code", icon: "claudecode-color", skills: `${HOME}/.claude/skills`, mcp: `${HOME}/.claude.json`, ...(on ? { hub: "on" } : {}) },
    { id: "codex", name: "Codex", icon: "codex-color", mcp: `${HOME}/.codex/config.toml`, ...(on ? { hub: "own" } : {}) },
    { id: "claude-desktop", name: "Claude Desktop", icon: "claude-color", mcp: `${HOME}/Library/Application Support/Claude/claude_desktop_config.json`, noRemote: true, ...(on ? { hub: "no-remote" } : {}) },
    // given none of the library's servers yet: Connect gives it them
    ...(st.cursor ? [{ id: "cursor", name: "Cursor", icon: "cursor", mcp: `${HOME}/.cursor/mcp.json`, ...(on && st.cursorOn ? { hub: "on" } : {}) }] : []),
  ],
  instructions: { agents: [], sets: [] }, foundServers: [], projects: [], foundSkills: [], skills: [], problems: [],
  servers: [
    { name: "my-tools", transport: "stdio", command: "node", args: [`${HOME}/code/my-tools/index.js`], agents: ["claude", "codex", "claude-desktop", ...(st.cursorOn ? ["cursor"] : [])] },
    { name: "notes", transport: "stdio", command: "uvx", args: ["notes-mcp"], agents: ["claude", "codex", ...(st.cursorOn ? ["cursor"] : [])] },
  ],
});

let I18N;
function tr(lang, key, vars = {}) {
  let s = key;
  if (lang !== "en") {
    assert.equal(typeof I18N[lang][key], "string", `${lang} has no ${JSON.stringify(key)}`);
    s = I18N[lang][key];
  }
  return s.replace(/\{(\w+)\}/g, (m, k) => (k in vars ? String(vars[k]) : m));
}

function server(lang, st) {
  return async (route) => {
    const req = route.request(), url = new URL(req.url());
    const json = (data) => route.fulfill({ json: data });
    if (url.pathname === "/boot.js") return route.fulfill({ contentType: "text/javascript", body: `window.bootPrefs = {lang:"${lang}",theme:"light",web:${st.web}};` });
    if (url.pathname === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
    if (url.pathname === "/api/state") return json({ agents: [], profiles: [], settings: { lang, theme: "light", gateway: "http://127.0.0.1:3425", publicURL: "https://mac.tail0.ts.net" } });
    if (url.pathname === "/api/library") return json(fixture(st.on, st));
    if (url.pathname === "/api/library/servers/hub") {
      const body = req.postDataJSON();
      st.sent.push(body);
      st.on = body.on;
      return json({ ...fixture(st.on, st), result: { changed: ["claude"] } });
    }
    if (url.pathname === "/api/library/servers/hub-open") {
      st.opened.push(req.postDataJSON());
      return json(fixture(st.on, st));
    }
    if (url.pathname === "/api/library/servers/agents-all") {
      st.all.push(req.postDataJSON());
      st.cursorOn = true;
      return json({ ...fixture(st.on, st), result: { changed: ["cursor"] } });
    }
    if (url.pathname === "/api/library/mcp/check") return json({ servers: Object.fromEntries(req.postDataJSON().names.map((n) => [n, { state: "ok", tools: 2 }])) });
    if (url.pathname === "/api/plugins") return json({ plugins: [] });
    if (url.pathname === "/api/usage/quotas") return json([]);
    if (url.pathname === "/api/groups") return json({ groups: [] });
    if (url.pathname === "/api/providers") return json({ providers: [], gateway: { running: true } });
    if (url.pathname.startsWith("/api/")) return json({});
    if (url.host !== "magpie.test") return route.fulfill({ status: 404, body: "" });
    const file = path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
    const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    try { await route.fulfill({ body: await fs.readFile(file), contentType }); } catch { await route.fulfill({ status: 404, body: "" }); }
  };
}

const sideways = (page) => page.evaluate(() => {
  const over = [];
  const w = document.documentElement.clientWidth;
  if (document.documentElement.scrollWidth > w) over.push("page " + document.documentElement.scrollWidth);
  for (const e of document.querySelectorAll("#view-library .lib-hub, #view-library .lib-hub .lib-switch, #view-library .lib-hub-why, #view-library .lib-hubat, #view-library .lib-hubat .lib-hubconnect, #view-library .lib-hubat .copy, .lib-hubsheet, .lib-hubsheet .segs, .lib-hubsheet .lib-hubcode, .lib-hubsheet .lib-hubagent, .lib-hubsheet .lib-hubagent .lib-hubgo, .lib-hubsheet .lib-hubok, .lib-hubsheet .bar button")) {
    const r = e.getBoundingClientRect();
    if (r.width && (r.right > w + 0.5 || r.left < -0.5)) over.push((e.className || e.tagName) + " " + Math.round(r.left) + "–" + Math.round(r.right));
  }
  return over;
});

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"])) {
  test(engine + ": one MCP server for every agent", async (t) => {
    assert(["chromium", "webkit"].includes(engine), "BROWSER must be chromium or webkit");
    I18N = strings().I18N;
    const browser = await (engine === "webkit" ? webkit.launch() : chromium.launch({ channel: process.env.PLAYWRIGHT_CHANNEL || "chromium" }));
    t.after(() => browser.close());

    for (const width of [1100, 420]) for (const lang of ["en", "zh", "zh-TW", "ja", "de"]) {
      await t.test(`${lang} ${width}px`, async () => {
        const w = (key, vars) => tr(lang, key, vars);
        const web = width < 800;
        // Cursor is one of the agents in the app, and not in a browser
        const st = { on: false, sent: [], opened: [], all: [], web, cursor: !web };
        const url = (web ? "https://mac.tail0.ts.net" : "http://127.0.0.1:3425") + "/mcp/magpie/hub";
        const errors = [];
        const ctx = await browser.newContext({ viewport: { width, height: 900 }, reducedMotion: "reduce" });
        await ctx.addInitScript(() => { try { localStorage.setItem("magpie.libTab", "mcp"); } catch {} });
        const page = await ctx.newPage();
        page.setDefaultTimeout(5000);
        page.on("pageerror", (e) => errors.push(e.message));
        await page.route("**/*", server(lang, st));
        await page.goto("http://magpie.test/");
        await page.locator('button[data-view="library"]').click();
        const v = page.locator("#view-library");
        const hub = v.locator(".lib-hub");
        await hub.waitFor();

        // off: what it would do, and no reasons yet
        assert.equal((await hub.locator(".name").textContent()).trim(), w("MCP hub"));
        assert.equal((await hub.locator(".sub.wraps").textContent()).trim(), w("Serve every server through magpie: each agent is given one server, magpie, in place of these, and any other app connects at one address."));
        assert.equal(await v.locator(".lib-hubstats, .lib-hubat").count(), 0);
        const sw = hub.locator(".lib-switch");
        assert.equal(await sw.getAttribute("role"), "switch");
        assert.equal(await sw.getAttribute("aria-checked"), "false");
        assert.equal(await v.locator(".lib-hub-why").count(), 0);
        // below the servers, not among them
        const below = await page.evaluate(() => {
          const rows = [...document.querySelectorAll("#view-library .lib-server")];
          const hub = document.querySelector("#view-library .lib-hub");
          return rows.length > 0 && rows.every((r) => r.getBoundingClientRect().bottom <= hub.getBoundingClientRect().top);
        });
        assert.ok(below, "the hub's row isn't below the servers");
        assert.deepEqual(await sideways(page), [], "sideways scroll, off");

        // on
        await sw.click();
        await page.waitForFunction((x) => document.querySelector("#status")?.textContent === x, w("The agents are given magpie's one MCP server — restart them to use it"));
        assert.deepEqual(st.sent, [{ on: true }]);
        await page.waitForFunction(() => document.querySelector("#view-library .lib-hub .lib-switch")?.getAttribute("aria-checked") === "true");
        assert.equal((await v.locator(".lib-hub .sub.wraps").textContent()).trim(), w("Each agent is given one server, magpie, in place of these. magpie runs each server once for all of them and hands each call to it, as server__tool. Restart an agent for it to read the change."));
        // what it serves: two servers, 2 tools each once checked, in Claude Code
        await page.waitForFunction((x) => document.querySelector("#view-library .lib-hubstats")?.textContent === x,
          [w("{n} servers", { n: 2 }), w("{n} tools", { n: 4 }), w("in {agents}", { agents: "Claude Code" })].join(" · "));
        // the address any other app connects at
        const at = v.locator(".lib-hubat");
        assert.equal((await at.locator(".name").textContent()).trim(), w("Other apps"));
        assert.equal((await at.locator(".lib-hubURL").textContent()).trim(), url);
        assert.equal((await at.locator(".lib-hubconnect").textContent()).trim(), w("Connect…"));
        const why = v.locator(".lib-hub-why");
        assert.deepEqual(await why.evaluateAll((ps) => ps.map((p) => p.dataset.why)), ["no-remote", "own"]);
        assert.equal((await v.locator('.lib-hub-why[data-why="no-remote"]').textContent()).trim(), w("{agents} reaches no MCP server over HTTP, so it keeps each server's own entry.", { agents: "Claude Desktop" }));
        assert.equal((await v.locator('.lib-hub-why[data-why="own"]').textContent()).trim(), w("{agents} has an MCP server called magpie of its own, which magpie leaves as it is, so it keeps each server's own entry.", { agents: "Codex" }));
        // the switch sits beside its text, inside the row
        const fits = await page.evaluate(() => {
          const row = document.querySelector("#view-library .lib-hub").getBoundingClientRect();
          const s = document.querySelector("#view-library .lib-hub .lib-switch").getBoundingClientRect();
          return s.width > 0 && s.left >= row.left && s.right <= row.right + 0.5 && s.top >= row.top && s.bottom <= row.bottom + 0.5;
        });
        assert.ok(fits, "the switch is outside its row");
        assert.deepEqual(await sideways(page), [], "sideways scroll, on");

        // Connect…: every agent here and where it stands, then the address
        // and what adds it, app by app
        await at.locator(".lib-hubconnect").click();
        const sheet = page.locator(".lib-hubsheet");
        await sheet.waitFor();
        assert.equal((await sheet.locator(".ehead b").textContent()).trim(), w("Connect an app"));
        assert.equal((await sheet.locator(".lib-hublead").textContent()).trim(), w("magpie connects the agents on this computer itself. Any other app that speaks MCP connects at the address below."));
        assert.deepEqual((await sheet.locator(".lib-hubsec").allTextContents()).map((x) => x.trim()), [w("Agents magpie looks after"), w("Any other app")]);
        const rows = sheet.locator(".lib-hubagent");
        const states = () => rows.evaluateAll((rs) => rs.map((r) => [r.dataset.agent, r.dataset.state]));
        assert.deepEqual(await states(), [["claude", "on"], ["codex", "own"], ["claude-desktop", "no-remote"], ...(web ? [] : [["cursor", "none"]])]);
        const line = (id) => sheet.locator(`.lib-hubagent[data-agent="${id}"] .sub`).textContent().then((x) => x.trim());
        assert.equal(await line("claude"), w("{n} servers, through magpie", { n: 2 }));
        assert.equal((await sheet.locator('.lib-hubagent[data-agent="claude"] .lib-hubok').textContent()).trim(), w("On the hub"));
        assert.equal(await line("codex"), w("It has a server called magpie of its own, which magpie leaves as it is"));
        assert.equal(await line("claude-desktop"), w("It reaches no MCP server over HTTP, so it keeps each server's own entry"));
        // only the one with nothing yet can be connected, and nothing is a command to run
        assert.equal(await sheet.locator(".lib-hubgo").count(), web ? 0 : 1);
        assert.equal(await sheet.locator(".lib-hubok").count(), 1);
        assert.doesNotMatch(await sheet.textContent(), /mcp add/);
        if (!web) {
          assert.equal(await line("cursor"), w("None of the library's servers is on for it yet"));
          const go = sheet.locator('.lib-hubagent[data-agent="cursor"] .lib-hubgo');
          assert.equal((await go.textContent()).trim(), w("Connect"));
          await go.click();
          await page.waitForFunction((x) => document.querySelector("#status")?.textContent === x, w("{agent} is connected — restart it to use magpie's tools", { agent: "Cursor" }));
          assert.deepEqual(st.all, [{ agents: ["cursor"], on: true }]);
          await page.waitForFunction(() => document.querySelector('.lib-hubagent[data-agent="cursor"]')?.dataset.state === "on");
          assert.equal(await line("cursor"), w("{n} servers, through magpie", { n: 2 }));
          assert.equal(await sheet.locator(".lib-hubgo").count(), 0);
          // the sheet stays open over the page drawn again
          assert.ok(await sheet.isVisible());
        }
        // each agent's icon beside its name, at every width
        const beside = await rows.evaluateAll((rs) => rs.filter((r) => {
          const i = r.querySelector(":scope > .ic").getBoundingClientRect(), n = r.querySelector(".name").getBoundingClientRect();
          return i.right > n.left || i.bottom < n.top || i.top > n.bottom;
        }).map((r) => r.dataset.agent));
        assert.deepEqual(beside, [], "an agent's icon isn't beside its name");
        assert.deepEqual(await sideways(page), [], "sideways scroll, agents");

        assert.equal((await sheet.locator(".lib-hubaddr code").textContent()).trim(), url);
        const opts = sheet.locator(".segs .opt");
        // Cursor is set up by hand only when magpie doesn't look after it
        assert.deepEqual((await opts.allTextContents()).map((x) => x.trim()), [...(web ? ["Cursor"] : []), "VS Code", w("Other")]);
        const pane = sheet.locator(".lib-hubpane");
        const code = async () => JSON.parse(await pane.locator(".lib-hubcode pre").textContent());
        if (web) {
          assert.equal((await pane.locator(".lib-hubwhere").textContent()).trim(), w("Or add it to ~/.cursor/mcp.json:"));
          assert.deepEqual(await code(), { mcpServers: { magpie: { url } } });
          assert.equal(await pane.locator(".lib-hubadd").count(), 0);
          await opts.nth(1).click();
        }
        // VS Code: its install link in the app, its file either way
        assert.equal((await pane.locator(".lib-hubwhere").textContent()).trim(), w("Or add it to .vscode/mcp.json, or to your user mcp.json:"));
        assert.deepEqual(await code(), { servers: { magpie: { type: "http", url } } });
        if (web) assert.equal(await pane.locator(".lib-hubadd").count(), 0);
        else {
          assert.equal((await pane.locator(".lib-hubadd").textContent()).trim(), w("Add to {app}", { app: "VS Code" }));
          await pane.locator(".lib-hubadd").click();
          await page.waitForFunction((x) => document.querySelector("#status")?.textContent === x, w("{app} asks you to confirm adding magpie", { app: "VS Code" }));
          assert.deepEqual(st.opened, [{ app: "vscode" }]);
        }
        assert.deepEqual(await sideways(page), [], "sideways scroll, VS Code");
        await opts.last().click();
        assert.deepEqual(await code(), { mcpServers: { magpie: { type: "http", url } } });
        assert.equal((await pane.locator(".lib-aside").textContent()).trim(), w("Its transport is Streamable HTTP. From another computer, use the address magpie is reached at there, with a gateway key: Authorization: Bearer <key>."));
        assert.deepEqual(await sideways(page), [], "sideways scroll, Other");
        assert.deepEqual(st.opened, web ? [] : [{ app: "vscode" }]);
        await sheet.locator(".bar button").click();
        await sheet.waitFor({ state: "hidden" });

        // off again
        await v.locator(".lib-hub .lib-switch").click();
        await page.waitForFunction((x) => document.querySelector("#status")?.textContent === x, w("The agents are given each server's own entry again — restart them to use it"));
        assert.deepEqual(st.sent, [{ on: true }, { on: false }]);
        await page.waitForFunction(() => document.querySelector("#view-library .lib-hub .lib-switch")?.getAttribute("aria-checked") === "false");
        assert.equal(await v.locator(".lib-hub-why").count(), 0);
        assert.equal(await v.locator(".lib-hubstats, .lib-hubat").count(), 0);

        assert.deepEqual(errors, []);
        await ctx.close();
      });
    }
  });
}
