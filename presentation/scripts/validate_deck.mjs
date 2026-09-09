import { createRequire } from "node:module";
import { join } from "node:path";

const require = createRequire(import.meta.url);
const modules = process.env.WORKSPACE_NODE_MODULES;
if (!modules) throw new Error("set WORKSPACE_NODE_MODULES to the bundled node_modules directory");
const { chromium } = require(join(modules, "playwright"));

const chrome = process.env.CHROME_PATH || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const baseUrl = process.argv[2] || "http://localhost:8080";
const viewports = [
  { width: 320, height: 568 },
  { width: 768, height: 1024 },
  { width: 1440, height: 900 },
];

const failures = [];
const browser = await chromium.launch({ headless: true, executablePath: chrome });

for (const viewport of viewports) {
  const context = await browser.newContext({ viewport });
  const page = await context.newPage();
  const browserErrors = [];
  page.on("pageerror", error => browserErrors.push(error.message));
  page.on("console", message => {
    if (message.type() === "error") browserErrors.push(message.text());
  });
  await page.goto(`${baseUrl}/#10`, { waitUntil: "networkidle" });
  const result = await page.evaluate(() => {
    const ids = [...document.querySelectorAll("[id]")].map(element => element.id);
    return {
      slides: document.querySelectorAll(".slide").length,
      notes: document.querySelectorAll(".slide .slide-notes").length,
      duplicateIds: ids.filter((id, index) => ids.indexOf(id) !== index),
      brokenImages: [...document.images].filter(image => !image.complete || !image.naturalWidth).map(image => image.src),
      bodyOverflow: document.documentElement.scrollWidth > innerWidth,
      stageOverflow: document.getElementById("stage").getBoundingClientRect().width > innerWidth + 1,
    };
  });
  if (result.slides !== 18) failures.push(`${viewport.width}px: expected 18 slides, found ${result.slides}`);
  if (result.notes !== result.slides) failures.push(`${viewport.width}px: ${result.slides - result.notes} slides lack notes`);
  if (result.duplicateIds.length) failures.push(`${viewport.width}px: duplicate ids ${result.duplicateIds.join(", ")}`);
  if (result.brokenImages.length) failures.push(`${viewport.width}px: broken images ${result.brokenImages.join(", ")}`);
  if (result.bodyOverflow || result.stageOverflow) failures.push(`${viewport.width}px: horizontal overflow`);
  if (browserErrors.length) failures.push(`${viewport.width}px: browser errors ${browserErrors.join(" | ")}`);
  await context.close();
}

const reducedContext = await browser.newContext({ viewport: { width: 1440, height: 900 }, reducedMotion: "reduce" });
const reducedPage = await reducedContext.newPage();
await reducedPage.goto(`${baseUrl}/#10`, { waitUntil: "networkidle" });
const reduced = await reducedPage.evaluate(() => ({
  hiddenFragments: [...document.querySelectorAll(".slide.on .fragment")].filter(element => getComputedStyle(element).opacity === "0").length,
  finalCaptionVisible: getComputedStyle(document.querySelector(".slide.on .system-caption span:last-child")).opacity === "1",
}));
if (reduced.hiddenFragments) failures.push(`reduced motion: ${reduced.hiddenFragments} fragments remain hidden`);
if (!reduced.finalCaptionVisible) failures.push("reduced motion: final architecture caption is hidden");
await reducedContext.close();

const interactionContext = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const interactionPage = await interactionContext.newPage();
await interactionPage.goto(`${baseUrl}/#18`, { waitUntil: "networkidle" });
if ((await interactionPage.locator("#count").textContent()) !== "18 / 18") failures.push("deep link: #18 did not open the final slide");
await interactionPage.goto(`${baseUrl}/#10`, { waitUntil: "networkidle" });
await interactionPage.keyboard.press("ArrowRight");
if ((await interactionPage.locator("#beat-count").textContent()) !== "1/6") failures.push("keyboard: ArrowRight did not advance one architecture beat");
await interactionPage.keyboard.press("r");
if ((await interactionPage.locator("#beat-count").textContent()) !== "0/6") failures.push("replay: R did not reset the architecture");
await interactionPage.getByRole("button", { name: "Sources", exact: true }).click();
if (!(await interactionPage.locator("#sources").evaluate(element => element.classList.contains("on")))) failures.push("sources overlay did not open");
if ((await interactionPage.locator("#stage").getAttribute("data-theme")) !== "light") failures.push("fresh context did not use the light theme");
await interactionContext.close();

await browser.close();

if (failures.length) {
  console.error(failures.map(failure => `ERROR: ${failure}`).join("\n"));
  process.exit(1);
}
console.log("deck validation passed: 18 slides, responsive layouts, reduced motion, deep links, controls, sources, and light default");
