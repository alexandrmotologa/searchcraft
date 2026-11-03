import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import { Resvg } from '@resvg/resvg-js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function buildFalconLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <!-- Background Luxury Gradient -->
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#ffffff"/>
      <stop offset="100%" stop-color="#f8fafc"/>
    </linearGradient>

    <!-- Luxury Double-Layer Squircle Shadow -->
    <filter id="squircleShadow" x="-10%" y="-10%" width="130%" height="130%">
      <feDropShadow dx="0" dy="28" stdDeviation="36" flood-color="#090d16" flood-opacity="0.09"/>
      <feDropShadow dx="0" dy="6" stdDeviation="12" flood-color="#090d16" flood-opacity="0.05"/>
    </filter>

    <!-- Radar Hexagonal Ambient Glow -->
    <linearGradient id="radarAura" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#0284c7" stop-opacity="0.14"/>
      <stop offset="100%" stop-color="#0f172a" stop-opacity="0.02"/>
    </linearGradient>

    <!-- Volumetric Obsidian & Titanium Lighting Tones -->
    <linearGradient id="titaniumShine" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#e2e8f0"/>
      <stop offset="100%" stop-color="#94a3b8"/>
    </linearGradient>

    <linearGradient id="lightFacet" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#94a3b8"/>
      <stop offset="100%" stop-color="#64748b"/>
    </linearGradient>

    <linearGradient id="slateLight" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#64748b"/>
      <stop offset="100%" stop-color="#475569"/>
    </linearGradient>

    <linearGradient id="slateMid" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>

    <linearGradient id="slateDark" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#1e293b"/>
      <stop offset="100%" stop-color="#0f172a"/>
    </linearGradient>

    <linearGradient id="obsidianDeep" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#0f172a"/>
      <stop offset="100%" stop-color="#020617"/>
    </linearGradient>

    <!-- Golden Raptor Beak -->
    <linearGradient id="beakGoldLight" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24"/>
      <stop offset="100%" stop-color="#f59e0b"/>
    </linearGradient>
    <linearGradient id="beakGoldDark" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#d97706"/>
      <stop offset="100%" stop-color="#92400e"/>
    </linearGradient>

    <!-- Electric Cyan Predator Optics -->
    <linearGradient id="cyanHotspot" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#ffffff"/>
      <stop offset="40%" stop-color="#38bdf8"/>
      <stop offset="100%" stop-color="#0284c7"/>
    </linearGradient>

    <filter id="eyeGlow" x="-30%" y="-30%" width="160%" height="160%">
      <feDropShadow dx="0" dy="0" stdDeviation="12" flood-color="#00f5ff" flood-opacity="0.85"/>
    </filter>
  </defs>

  <!-- 1. Luxury White Squircle Container -->
  <rect x="32" y="32" width="960" height="960" rx="220" fill="url(#bgGrad)" filter="url(#squircleShadow)" stroke="#e2e8f0" stroke-width="4"/>

  <!-- 2. Concentric Inner Framing Line -->
  <rect x="58" y="58" width="908" height="908" rx="194" fill="none" stroke="#f1f5f9" stroke-width="2"/>

  <!-- 3. Hexagonal Architectural Gateway -->
  <polygon points="512,95 835,250 835,650 512,930 189,650 189,250"
           fill="url(#radarAura)" stroke="#cbd5e1" stroke-width="3" stroke-dasharray="14 10"/>
  <polygon points="512,120 810,265 810,635 512,905 214,635 214,265"
           fill="none" stroke="#0284c7" stroke-width="2" stroke-opacity="0.4"/>

  <!-- 4. Watertight Solid Falcon Base Silhouette -->
  <path d="M 512,135
           L 590,165 L 685,225 L 760,335 L 795,475 L 760,635 L 795,750 L 705,875 L 512,915
           L 319,875 L 229,750 L 264,635 L 229,475 L 264,335 L 339,225 L 434,165 Z"
        fill="#020617"/>

  <!-- 5. Low-Poly Volumetric Facets -->

  <!-- CROWN & FOREHEAD CREST -->
  <polygon points="512,135 434,165 470,240 512,230" fill="url(#titaniumShine)"/>
  <polygon points="512,135 590,165 554,240 512,230" fill="url(#lightFacet)"/>
  <polygon points="434,165 339,225 410,270 470,240" fill="url(#lightFacet)"/>
  <polygon points="590,165 685,225 614,270 554,240" fill="url(#slateLight)"/>

  <!-- TEMPLES & SUPERCILIARY RIDGE (High Alert Angles) -->
  <polygon points="339,225 264,335 380,330 410,270" fill="url(#slateLight)"/>
  <polygon points="685,225 760,335 644,330 614,270" fill="url(#obsidianDeep)"/>

  <!-- AGGRESSIVE V-CHEVRON RADAR BROW -->
  <polygon points="512,230 470,240 410,270 465,320 512,330" fill="url(#titaniumShine)"/>
  <polygon points="512,230 554,240 614,270 559,320 512,330" fill="url(#slateMid)"/>
  <polygon points="410,270 380,330 440,345 465,320" fill="url(#lightFacet)"/>
  <polygon points="614,270 644,330 584,345 559,320" fill="url(#slateDark)"/>

  <!-- NASAL BRIDGE (Between the Eyes) -->
  <polygon points="512,330 465,320 470,410 512,415" fill="url(#titaniumShine)"/>
  <polygon points="512,330 559,320 554,410 512,415" fill="url(#slateMid)"/>

  <!-- PREDATOR EYE SOCKETS & ACUTE ALMOND EYES (UP & OUT CANT) -->
  <!-- Left Eye Socket -->
  <polygon points="465,320 440,345 365,370 435,420 470,410" fill="#020617"/>
  <!-- Left Eye (Slanted Upward: Outer 395,355 - Inner 465,395) -->
  <polygon points="400,355 445,360 465,395 430,395" fill="url(#cyanHotspot)" filter="url(#eyeGlow)"/>
  <polygon points="415,362 445,368 458,390 435,388" fill="#ffffff"/>

  <!-- Right Eye Socket -->
  <polygon points="559,320 584,345 659,370 589,420 554,410" fill="#030712"/>
  <!-- Right Eye (Slanted Upward: Outer 624,355 - Inner 559,395) -->
  <polygon points="624,355 579,360 559,395 594,395" fill="url(#cyanHotspot)" filter="url(#eyeGlow)"/>
  <polygon points="609,362 579,368 566,390 589,388" fill="#ffffff"/>

  <!-- MALAR CHEEKS & INTEGRATED POLYGONAL MARKINGS -->
  <polygon points="380,330 264,335 229,475 320,490 365,370" fill="url(#slateLight)"/>
  <polygon points="644,330 760,335 795,475 704,490 659,370" fill="url(#obsidianDeep)"/>
  <polygon points="365,370 435,420 445,490 355,530 320,490" fill="url(#slateMid)"/>
  <polygon points="659,370 589,420 579,490 669,530 704,490" fill="url(#obsidianDeep)"/>

  <!-- CERE (Nasal Plate) -->
  <polygon points="512,415 470,410 460,485 512,495" fill="url(#titaniumShine)"/>
  <polygon points="512,415 554,410 564,485 512,495" fill="url(#slateMid)"/>
  <polygon points="470,410 435,420 445,490 460,485" fill="url(#lightFacet)"/>
  <polygon points="554,410 589,420 579,490 564,485" fill="url(#slateDark)"/>

  <!-- COMPACT HOOKED FALCON BEAK (Amber Gold) -->
  <polygon points="512,495 460,485 470,575 512,640" fill="url(#beakGoldLight)"/>
  <polygon points="512,495 564,485 554,575 512,640" fill="url(#beakGoldDark)"/>
  <polygon points="470,575 512,640 495,660" fill="#f59e0b"/>
  <polygon points="554,575 512,640 512,660" fill="#b45309"/>

  <!-- JAW & LOWER CHEST -->
  <polygon points="460,485 445,490 420,580 470,575" fill="url(#slateMid)"/>
  <polygon points="564,485 579,490 604,580 554,575" fill="url(#obsidianDeep)"/>
  <polygon points="512,640 470,575 420,580 475,665 512,695" fill="url(#slateMid)"/>
  <polygon points="512,640 554,575 604,580 549,665 512,695" fill="url(#slateDark)"/>

  <!-- BREASTPLATE & DENSE VECTOR RADAR CHEVRONS -->
  <polygon points="512,695 475,665 410,650 430,765 512,825" fill="url(#slateLight)"/>
  <polygon points="512,695 549,665 614,650 594,765 512,825" fill="url(#slateDark)"/>
  <polygon points="512,825 430,765 450,875 512,915" fill="url(#slateMid)"/>
  <polygon points="512,825 594,765 574,875 512,915" fill="url(#obsidianDeep)"/>

  <!-- WINGS & FLANK MANTLE -->
  <polygon points="229,475 264,635 355,530 320,490" fill="url(#slateMid)"/>
  <polygon points="795,475 760,635 669,530 704,490" fill="url(#obsidianDeep)"/>
  <polygon points="264,635 229,750 335,730 355,530" fill="url(#slateLight)"/>
  <polygon points="760,635 795,750 689,730 669,530" fill="url(#obsidianDeep)"/>
  <polygon points="355,530 335,730 410,650 420,580" fill="url(#slateMid)"/>
  <polygon points="669,530 689,730 614,650 604,580" fill="url(#obsidianDeep)"/>
  <polygon points="229,750 319,875 450,875 430,765 335,730" fill="url(#slateDark)"/>
  <polygon points="795,750 705,875 574,875 594,765 689,730" fill="#030712"/>
  <polygon points="319,875 512,915 450,875" fill="url(#slateMid)"/>
  <polygon points="705,875 512,915 574,875" fill="url(#obsidianDeep)"/>
</svg>`;
}

async function main() {
  const outputDir = path.resolve(__dirname, '../docs/images');
  fs.mkdirSync(outputDir, { recursive: true });

  const svg = buildFalconLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf-8');

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: 1024 },
  });
  const pngBuffer = resvg.render().asPng();
  fs.writeFileSync(pngPath, pngBuffer);

  console.log(`[Developer Branding] Successfully generated predatory falcon logo:`);
  console.log(` -> ${svgPath}`);
  console.log(` -> ${pngPath} (1024x1024)`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
