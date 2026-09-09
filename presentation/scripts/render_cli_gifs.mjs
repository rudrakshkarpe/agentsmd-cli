import { createRequire } from "node:module";
import { cp, mkdir, rm } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const require = createRequire(import.meta.url);
const modules = process.env.WORKSPACE_NODE_MODULES;
if (!modules) throw new Error("set WORKSPACE_NODE_MODULES to the bundled node_modules directory");
const { chromium } = require(join(modules, "playwright"));

const here = dirname(fileURLToPath(import.meta.url));
const presentation = resolve(here, "..");
const assets = join(presentation, "assets");
const cache = join(presentation, ".render-cache");
const chrome = process.env.CHROME_PATH || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const ffmpeg = process.env.FFMPEG_PATH || "/opt/homebrew/bin/ffmpeg";

const pageHtml = `<!doctype html><html><head><meta charset="utf-8"><style>
*{box-sizing:border-box}html,body{margin:0;width:100%;height:100%;overflow:hidden;background:#050609;color:#e8edf4;font-family:"IBM Plex Mono",Menlo,monospace}.canvas{width:1280px;height:720px;display:grid;place-items:center;background:radial-gradient(circle at 80% 15%,#10233d 0,transparent 36%),#050609}.terminal{width:1160px;height:624px;overflow:hidden;border:1px solid #303640;border-radius:15px;background:#0e1014;box-shadow:0 30px 90px #0009}.head{height:52px;display:flex;align-items:center;gap:9px;padding:0 18px;border-bottom:1px solid #303640;background:#181b21}.dot{width:11px;height:11px;border-radius:50%;background:#59616e}.title{margin-left:10px;color:#8993a1;font-size:15px;font-weight:600}.badge{margin-left:auto;color:#74adff;font-size:12px;letter-spacing:.1em;text-transform:uppercase}.screen{height:572px;margin:0;padding:22px 26px;overflow:hidden;font-size:18px;line-height:1.42;white-space:pre-wrap}.line{min-height:25px}.brand{color:#6ee7a8;font-weight:600}.prompt{color:#77afff}.ok{color:#6ee7a8}.warn{color:#ffc66d}.dim{color:#8993a1}.soft{color:#d7dde7}.rule{color:#9ee8bd;background:#123525;padding:1px 6px}.cursor{display:inline-block;width:9px;height:20px;margin-left:3px;background:#77afff;vertical-align:-3px;animation:blink 1s steps(1) infinite}@keyframes blink{50%{opacity:0}}</style></head><body><div class="canvas"><div class="terminal"><div class="head"><i class="dot"></i><i class="dot"></i><i class="dot"></i><span class="title" id="title"></span><span class="badge">recorded CLI</span></div><div class="screen" id="screen"></div></div></div></body></html>`;

const setup = {
  name: "agentsmd-setup",
  title: "agentsmd · repository setup",
  intro: [
    " ███   ████ █████ █   █ █████  ████ █   █ ████         ████ █     █████",
    "█   █ █     █     ██  █   █   █     ██ ██ █   █       █     █       █",
    "█████ █  ██ ████  █ █ █   █    ███  █ █ █ █   █ █████ █     █       █",
    "█   █ █   █ █     █  ██   █       █ █   █ █   █       █     █       █",
    "█   █  ███  █████ █   █   █   ████  █   █ ████         ████ █████ █████",
    "",
    "Project-aware guidance for coding agents · agentsmd-cli"
  ],
  actions: [
    { command: "agentsmd init", output: [["✓ initialized ~/work/envmerge (v0000)", "ok"], ["ℹ Detected: Go · 2 verified command(s)", "dim"], ["ℹ Next: review AGENTS.md, run `agentsmd doctor`, then connect your coding CLI.", "dim"]] },
    { command: "agentsmd doctor", output: [["🩺 agentsmd CLI · doctor", "brand"], ["✓ Git — /usr/bin/git", "ok"], ["✓ AGENTS.md — ~/work/envmerge/AGENTS.md", "ok"], ["! Automation — reflection disabled", "warn"], ["✓ Queue — healthy", "ok"], ["! Codex — ~/.local/bin/codex · ready to connect", "warn"]] },
    { command: "agentsmd connect codex", output: [["✓ connected Codex", "ok"], ["ℹ ~/work/envmerge/.codex/hooks.json", "dim"]] },
    { command: "agentsmd automate", output: [["🧠 agentsmd CLI · automation", "brand"], ["Reflection  disabled", "dim"], ["Evaluation  disabled", "dim"], ["Promotion   manual", "soft"], ["Confidence  0.80", "soft"]] }
  ]
};

const learning = {
  name: "agentsmd-learning",
  title: "agentsmd · config-precedence task",
  actions: [
    { command: "agentsmd task start config-precedence", output: [["✓ active task config-precedence", "ok"]] },
    { command: "agentsmd sessions", output: [["📡 agentsmd CLI · sessions", "brand"], ["Recorded agent work and evaluation outcomes", "dim"], ["✓ baseline-1  codex · success   30.0s  files 2  tests 1/0", "ok"], ["✓ learned-1   codex · success   20.0s  files 2  tests 1/0", "ok"]] },
    { command: "agentsmd learn --task config-precedence --run baseline-1 --rule \"Keep GOCACHE inside restricted workspaces.\"", output: [["✓ proposed p1788976929754666000", "ok"]] },
    { command: "agentsmd pending", output: [["⏳ p1788976929754666000  task=config-precedence", "warn"], ["Keep GOCACHE inside restricted workspaces.", "soft"]] },
    { command: "agentsmd promote p1788976929754666000", output: [["✓ promoted r000", "ok"]] },
    { command: "agentsmd blame", output: [["[r000] run=baseline-1 task=config-precedence cited=0", "dim"], ["Keep GOCACHE inside restricted workspaces.", "rule"]] }
  ]
};

async function renderGif(browser, demo) {
  const framesDir = join(cache, demo.name);
  await rm(framesDir, { recursive: true, force: true });
  await mkdir(framesDir, { recursive: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 }, deviceScaleFactor: 1 });
  await page.setContent(pageHtml);
  await page.locator("#title").evaluate((element, text) => { element.textContent = text; }, demo.title);
  let frame = 0;

  const snap = async (copies = 1) => {
    const first = join(framesDir, `frame-${String(frame++).padStart(4, "0")}.png`);
    await page.screenshot({ path: first });
    for (let i = 1; i < copies; i++) {
      await cp(first, join(framesDir, `frame-${String(frame++).padStart(4, "0")}.png`));
    }
  };
  const clear = () => page.locator("#screen").evaluate(element => { element.replaceChildren(); });
  const addLine = (text, tone = "soft") => page.locator("#screen").evaluate((element, value) => {
    const line = document.createElement("div");
    line.className = `line ${value.tone}`;
    line.textContent = value.text;
    element.appendChild(line);
    element.scrollTop = element.scrollHeight;
  }, { text, tone });
  const typeCommand = async command => {
    await page.locator("#screen").evaluate(element => {
      const line = document.createElement("div");
      line.className = "line prompt";
      line.innerHTML = '<span>$ </span><span class="typed"></span><i class="cursor"></i>';
      element.appendChild(line);
      element.scrollTop = element.scrollHeight;
    });
    for (let index = 0; index <= command.length; index += 4) {
      await page.locator(".typed").last().evaluate((element, text) => { element.textContent = text; }, command.slice(0, Math.min(index, command.length)));
      await snap();
    }
    await page.locator(".cursor").last().evaluate(element => element.remove());
    await snap(2);
  };

  if (demo.intro) {
    for (const line of demo.intro) await addLine(line, line.startsWith(" ") || line.startsWith("█") ? "brand" : "soft");
    await snap(18);
    await clear();
  }
  for (const [index, action] of demo.actions.entries()) {
    if (index && await page.locator("#screen .line").count() > 17) await clear();
    await typeCommand(action.command);
    for (const [text, tone] of action.output) {
      await addLine(text, tone);
      await snap(2);
    }
    await addLine("");
    await snap(3);
  }
  await snap(20);
  await page.close();

  const output = join(assets, `${demo.name}.gif`);
  const filter = "fps=8,scale=960:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128:stats_mode=diff[p];[s1][p]paletteuse=dither=bayer:bayer_scale=3:diff_mode=rectangle";
  const result = spawnSync(ffmpeg, ["-y", "-framerate", "8", "-i", join(framesDir, "frame-%04d.png"), "-filter_complex", filter, "-loop", "0", output], { stdio: "inherit" });
  if (result.status !== 0) throw new Error(`ffmpeg failed for ${demo.name}`);
  await rm(framesDir, { recursive: true, force: true });
}

await mkdir(assets, { recursive: true });
await mkdir(cache, { recursive: true });
const browser = await chromium.launch({ headless: true, executablePath: chrome });
try {
  await renderGif(browser, setup);
  await renderGif(browser, learning);
} finally {
  await browser.close();
  await rm(cache, { recursive: true, force: true });
}
