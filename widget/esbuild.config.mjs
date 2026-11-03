import * as esbuild from 'esbuild';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const isWatch = process.argv.includes('--watch');
const outDir = path.resolve(__dirname, 'dist');
const distInServer = path.resolve(__dirname, '../internal/server/dist/widget');

if (!fs.existsSync(outDir)) {
  fs.mkdirSync(outDir, { recursive: true });
}
if (!fs.existsSync(distInServer)) {
  fs.mkdirSync(distInServer, { recursive: true });
}

// Read and minify CSS
const cssRaw = fs.readFileSync(path.resolve(__dirname, 'src/styles.css'), 'utf-8');
const minifiedCss = cssRaw
  .replace(/\/\*[\s\S]*?\*\//g, '')
  .replace(/\s+/g, ' ')
  .replace(/\s*([\{\}:;,])\s*/g, '$1')
  .trim();

// Code to inject CSS into head
const cssInjector = `
(function() {
  if (typeof document !== 'undefined' && !document.getElementById('sc-widget-styles')) {
    var style = document.createElement('style');
    style.id = 'sc-widget-styles';
    style.textContent = ${JSON.stringify(minifiedCss)};
    (document.head || document.documentElement).appendChild(style);
  }
})();
`;

async function build() {
  const result = await esbuild.build({
    entryPoints: [path.resolve(__dirname, 'src/index.ts')],
    bundle: true,
    minify: true,
    target: ['es2020'],
    format: 'iife',
    write: false,
    banner: {
      js: cssInjector,
    },
  });

  const outputJs = result.outputFiles[0].text;
  const outPath = path.resolve(outDir, 'widget.js');
  const serverPath = path.resolve(distInServer, 'widget.js');

  fs.writeFileSync(outPath, outputJs, 'utf-8');
  fs.writeFileSync(serverPath, outputJs, 'utf-8');

  const sizeKb = (Buffer.byteLength(outputJs, 'utf-8') / 1024).toFixed(2);
  console.log(`[SearchCraft Widget] Built standalone bundle: ${sizeKb} KB`);
  console.log(` -> ${outPath}`);
  console.log(` -> ${serverPath}`);
}

build().catch((err) => {
  console.error(err);
  process.exit(1);
});
