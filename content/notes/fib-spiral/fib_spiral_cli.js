#!/usr/bin/env node

"use strict";

function usage() {
  return [
    "Usage:",
    "  node gen_fib_spiral.js [--fib-segments <n>] [--fib-squares]",
    "",
    "Options:",
    "  --fib-segments <n>  Number of Fibonacci segments to generate (default: 14)",
    "  --fib-squares       Render squares + corresponding arcs",
    "  -h, --help          Show this help message",
  ].join("\n");
}

function fail(message) {
  process.stderr.write(`${message}\n\n${usage()}\n`);
  process.exit(1);
}

function parseArgs(argv) {
  let fibSegments = 14;
  let fibSquares = false;

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--fib-squares") {
      fibSquares = true;
      continue;
    }
    if (arg === "-h" || arg === "--help") {
      process.stdout.write(`${usage()}\n`);
      process.exit(0);
    }
    if (arg === "--fib-segments") {
      const value = argv[i + 1];
      if (value == null) {
        fail("Missing value for --fib-segments.");
      }
      const parsed = Number(value);
      if (!Number.isInteger(parsed) || parsed < 2) {
        fail("--fib-segments must be an integer >= 2.");
      }
      fibSegments = parsed;
      i += 1;
      continue;
    }
    if (arg.startsWith("--fib-segments=")) {
      const parsed = Number(arg.slice("--fib-segments=".length));
      if (!Number.isInteger(parsed) || parsed < 2) {
        fail("--fib-segments must be an integer >= 2.");
      }
      fibSegments = parsed;
      continue;
    }
    fail(`Unknown argument: ${arg}`);
  }

  return { fibSegments, fibSquares };
}

// Complex helpers (re + i*im)
const C = (re, im) => ({ re, im });
const add = (a, b) => C(a.re + b.re, a.im + b.im);
const mul = (a, b) => C(a.re * b.re - a.im * b.im, a.re * b.im + a.im * b.re);

// Rotations: -i, -1, +i, 1
const rots = [C(0, -1), C(-1, 0), C(0, 1), C(1, 0)];

function* fibs() {
  let a = 0;
  let b = 1;
  while (true) {
    yield a;
    const c = a + b;
    a = b;
    b = c;
  }
}

function collectFibSegments(n) {
  const segs = [];
  let acc = C(0, 0);

  const fibgen = fibs();
  fibgen.next();
  fibgen.next();

  for (let i = 0; i < n; i += 1) {
    const Fn = fibgen.next().value;
    const rot = rots[i % 4];

    segs.push({ Fn, P: acc, rot });

    const step = mul(rot, C(Fn, Fn));
    acc = add(acc, step);
  }

  return segs;
}

function boundsForFibGeometry(segs) {
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;

  for (let k = 0; k < segs.length - 1; k += 1) {
    const { Fn, P, rot } = segs[k];
    const ux = mul(rot, C(Fn, 0));
    const vy = mul(rot, C(0, Fn));
    const c0 = P;
    const c1 = add(P, ux);
    const c2 = add(c1, vy);
    const c3 = add(P, vy);

    const corners = [c0, c1, c2, c3];
    for (const c of corners) {
      minX = Math.min(minX, c.re);
      minY = Math.min(minY, c.im);
      maxX = Math.max(maxX, c.re);
      maxY = Math.max(maxY, c.im);
    }
  }

  return { minX, minY, maxX, maxY };
}

function buildArcPath(P0, P1, Fn, dx, dy) {
  const x0 = P0.re + dx;
  const y0 = P0.im + dy;
  const x1 = P1.re + dx;
  const y1 = P1.im + dy;
  return `<path class="arc" d="M ${x0} ${y0} A ${Fn} ${Fn} 0 0 0 ${x1} ${y1}" />`;
}

function buildSquarePolygon(P, Fn, rot, dx, dy) {
  const ux = mul(rot, C(Fn, 0));
  const vy = mul(rot, C(0, Fn));
  const c0 = P;
  const c1 = add(P, ux);
  const c2 = add(c1, vy);
  const c3 = add(P, vy);
  const points = [c0, c1, c2, c3]
    .map((c) => `${c.re + dx},${c.im + dy}`)
    .join(" ");
  return `<polygon class="square" points="${points}" />`;
}

function renderSvg({ fibSegments, fibSquares }) {
  const segs = collectFibSegments(fibSegments);
  const bounds = boundsForFibGeometry(segs);

  const pad = 1;
  const width = Math.ceil(bounds.maxX - bounds.minX + 2 * pad);
  const height = Math.ceil(bounds.maxY - bounds.minY + 2 * pad);
  const dx = -bounds.minX + pad;
  const dy = -bounds.minY + pad;

  const shapes = [];

  if (fibSquares) {
    for (let k = 0; k < segs.length - 1; k += 1) {
      const { Fn, P, rot } = segs[k];
      shapes.push(buildSquarePolygon(P, Fn, rot, dx, dy));
    }
  }

  for (let k = 0; k < segs.length - 1; k += 1) {
    const { Fn, P } = segs[k];
    const Pn1 = segs[k + 1].P;
    shapes.push(buildArcPath(P, Pn1, Fn, dx, dy));
  }

  return [
    `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">`,
    "  <style>",
    "    .arc { fill: none; stroke: #bbb; stroke-width: 2; stroke-linecap: round; }",
    "    .square { fill: none; stroke: lime; stroke-width: 0.5; }",
    "  </style>",
    `<rect width="100%" height="100%" fill="#161616" />`,
    ...shapes.map((shape) => `  ${shape}`),
    "</svg>",
    "",
  ].join("\n");
}

const options = parseArgs(process.argv.slice(2));
process.stdout.write(renderSvg(options));
