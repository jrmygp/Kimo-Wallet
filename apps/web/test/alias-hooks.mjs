// Module resolution hooks for `npm test` (Node's built-in test runner with
// --experimental-strip-types). Node doesn't read tsconfig, so this maps the
// `@/…` alias to the app root and resolves extensionless .ts/.tsx imports the
// way the Next.js bundler does. Test-only; not used by the app build.
import { existsSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const appRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const candidates = ["", ".ts", ".tsx", "/index.ts", "/index.tsx"];

function resolveFile(absolutePath) {
  for (const suffix of candidates) {
    const candidate = absolutePath + suffix;
    if (existsSync(candidate) && statSync(candidate).isFile()) return candidate;
  }
  return null;
}

export async function resolve(specifier, context, nextResolve) {
  let target = null;
  if (specifier.startsWith("@/")) {
    target = resolveFile(path.join(appRoot, specifier.slice(2)));
  } else if ((specifier.startsWith("./") || specifier.startsWith("../")) && context.parentURL?.startsWith("file:")) {
    target = resolveFile(path.resolve(path.dirname(fileURLToPath(context.parentURL)), specifier));
  }
  return nextResolve(target ? pathToFileURL(target).href : specifier, context);
}
