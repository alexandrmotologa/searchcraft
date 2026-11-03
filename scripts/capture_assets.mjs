import puppeteer from 'puppeteer';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { execSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const outDir = path.resolve(__dirname, '../docs/images');
const framesDir = path.resolve(__dirname, '../docs/images/frames');

fs.mkdirSync(outDir, { recursive: true });
fs.mkdirSync(framesDir, { recursive: true });

async function sleep(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

async function main() {
  console.log('[Assets] Launching Chrome to capture authentic screenshots...');

  const browser = await puppeteer.launch({
    headless: 'new',
    executablePath: 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--force-device-scale-factor=2'],
    defaultViewport: { width: 1440, height: 900, deviceScaleFactor: 2 },
  });

  const page = await browser.newPage();

  // 1. Desktop Search Sandbox (Hybrid Search query)
  console.log('-> Capturing Search Sandbox (Desktop)...');
  await page.goto('http://127.0.0.1:7700', { waitUntil: 'networkidle0' });
  await sleep(600);

  // Type search query
  const inputSelector = 'input[type="text"]';
  await page.waitForSelector(inputSelector);
  await page.type(inputSelector, 'ergonomic wireless', { delay: 40 });
  await sleep(700);

  await page.screenshot({
    path: path.join(outDir, 'searchcraft_sandbox.png'),
    clip: { x: 0, y: 0, width: 1440, height: 900 },
  });
  console.log('✓ Saved searchcraft_sandbox.png');

  // 2. Query Analytics Tab
  console.log('-> Capturing Query Analytics...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('nav button'));
    const analyticsBtn = buttons.find((b) => b.textContent && b.textContent.includes('Analytics'));
    if (analyticsBtn) analyticsBtn.click();
  });
  await sleep(800);

  await page.screenshot({
    path: path.join(outDir, 'searchcraft_analytics.png'),
    clip: { x: 0, y: 0, width: 1440, height: 900 },
  });
  console.log('✓ Saved searchcraft_analytics.png');

  // 3. Document Explorer Tab
  console.log('-> Capturing Document Explorer...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('nav button'));
    const docsBtn = buttons.find((b) => b.textContent && b.textContent.includes('Documents'));
    if (docsBtn) docsBtn.click();
  });
  await sleep(800);

  await page.screenshot({
    path: path.join(outDir, 'searchcraft_documents.png'),
    clip: { x: 0, y: 0, width: 1440, height: 900 },
  });
  console.log('✓ Saved searchcraft_documents.png');

  // 4. Standalone Cmd+K Search Widget Modal
  console.log('-> Capturing 2KB Search Widget (Cmd+K modal)...');
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('nav button'));
    const widgetBtn = buttons.find((b) => b.textContent && b.textContent.includes('Widget'));
    if (widgetBtn) widgetBtn.click();
  });
  await sleep(700);

  // Click "Test Live Widget" button inside Widget Embed view
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const testBtn = buttons.find((b) => b.textContent && b.textContent.includes('Test Live Widget'));
    if (testBtn) {
      testBtn.click();
    } else if (window.SearchCraft) {
      window.SearchCraft.open();
    }
  });
  await sleep(600);

  // Type in widget modal
  const widgetInputSelector = '.sc-input';
  await page.waitForSelector(widgetInputSelector, { visible: true });
  await page.type(widgetInputSelector, 'mechanical keyboard', { delay: 40 });
  await sleep(600);

  await page.screenshot({
    path: path.join(outDir, 'searchcraft_widget.png'),
    clip: { x: 0, y: 0, width: 1440, height: 900 },
  });
  console.log('✓ Saved searchcraft_widget.png');

  // 5. Mobile Responsive View (390x844 iPhone viewport)
  console.log('-> Capturing Mobile Responsive View (390x844)...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2, isMobile: true });
  await page.goto('http://127.0.0.1:7700', { waitUntil: 'networkidle0' });
  await sleep(600);

  await page.waitForSelector(inputSelector);
  await page.type(inputSelector, 'keyboard', { delay: 30 });
  await sleep(600);

  await page.screenshot({
    path: path.join(outDir, 'searchcraft_mobile.png'),
    clip: { x: 0, y: 0, width: 390, height: 844 },
  });
  console.log('✓ Saved searchcraft_mobile.png');

  // 6. Generate Frames for Animated Demo GIF
  console.log('-> Generating frames for animated demo GIF...');
  await page.setViewport({ width: 1280, height: 800, deviceScaleFactor: 1.5 });
  await page.goto('http://127.0.0.1:7700', { waitUntil: 'networkidle0' });
  await sleep(600);

  let frameIdx = 0;
  async function captureFrame(repeat = 1) {
    const frameFile = path.join(framesDir, `frame_${String(frameIdx).padStart(3, '0')}.png`);
    await page.screenshot({
      path: frameFile,
      clip: { x: 0, y: 0, width: 1280, height: 800 },
    });
    frameIdx++;
    for (let i = 1; i < repeat; i++) {
      const extraFile = path.join(framesDir, `frame_${String(frameIdx).padStart(3, '0')}.png`);
      fs.copyFileSync(frameFile, extraFile);
      frameIdx++;
    }
  }

  // Frame 0-3: Initial view
  await captureFrame(4);

  // Frame 4-12: Type "ergonomic" character by character
  const chars = 'ergonomic'.split('');
  for (const c of chars) {
    await page.type(inputSelector, c, { delay: 60 });
    await sleep(80);
    await captureFrame(1);
  }
  await sleep(400);
  await captureFrame(5); // hold on "ergonomic"

  // Frame 13-18: Type " wireless"
  for (const c of ' wireless'.split('')) {
    await page.type(inputSelector, c, { delay: 60 });
    await sleep(70);
    await captureFrame(1);
  }
  await sleep(400);
  await captureFrame(6); // hold on results

  // Frame 19-24: Switch to Analytics tab
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('nav button'));
    const analyticsBtn = buttons.find((b) => b.textContent && b.textContent.includes('Analytics'));
    if (analyticsBtn) analyticsBtn.click();
  });
  await sleep(500);
  await captureFrame(8);

  // Frame 25-30: Switch to Widget tab & open Cmd+K modal
  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('nav button'));
    const widgetBtn = buttons.find((b) => b.textContent && b.textContent.includes('Widget'));
    if (widgetBtn) widgetBtn.click();
  });
  await sleep(400);
  await captureFrame(3);

  await page.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const testBtn = buttons.find((b) => b.textContent && b.textContent.includes('Test Live Widget'));
    if (testBtn) {
      testBtn.click();
    } else if (window.SearchCraft) {
      window.SearchCraft.open();
    }
  });
  await sleep(400);
  await captureFrame(2);

  await page.waitForSelector(widgetInputSelector, { visible: true });
  for (const c of 'headphone'.split('')) {
    await page.type(widgetInputSelector, c, { delay: 50 });
    await captureFrame(1);
  }
  await sleep(400);
  await captureFrame(8); // hold on widget results

  await browser.close();

  // 7. Compile GIF with ffmpeg
  console.log(`-> Compiling ${frameIdx} frames into searchcraft_demo.gif via ffmpeg...`);
  const gifOut = path.join(outDir, 'searchcraft_demo.gif');
  const ffmpegCmd = `ffmpeg -y -framerate 6 -i "${framesDir}/frame_%03d.png" -vf "scale=1000:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer" "${gifOut}"`;
  execSync(ffmpegCmd, { stdio: 'inherit' });

  // Cleanup frame pngs
  const files = fs.readdirSync(framesDir);
  for (const f of files) {
    fs.unlinkSync(path.join(framesDir, f));
  }
  fs.rmdirSync(framesDir);

  console.log(`✓ Successfully compiled animated demo GIF: ${gifOut}`);
  console.log('[Assets] All authentic screenshots and demo GIF generated successfully!');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
