import { readFileSync, writeFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const cssPath = resolve(__dirname, "../src/index.css");
const outputPath = resolve(__dirname, "../tokens.json");

const css = readFileSync(cssPath, "utf-8");

function extractBlock(css, selector) {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const re = new RegExp(`${escaped}\\s*\\{([^}]+)\\}`, "g");
  const match = re.exec(css);
  if (!match) return {};
  const vars = {};
  for (const line of match[1].split("\n")) {
    const m = line.match(/^\s*--([a-z0-9-]+):\s*(.+?)\s*;/);
    if (m) vars[m[1]] = m[2];
  }
  return vars;
}

const lightVars = extractBlock(css, ":root");
const darkVars = extractBlock(css, ".dark");

function categorize(name) {
  if (["background", "foreground", "card", "card-foreground", "popover", "popover-foreground",
       "primary", "primary-foreground", "secondary", "secondary-foreground",
       "muted", "muted-foreground", "accent", "accent-foreground",
       "destructive", "destructive-foreground", "border", "input", "ring",
       "success", "success-foreground", "warning", "warning-foreground",
       "danger", "danger-foreground", "surface", "surface-foreground"].includes(name)) return "color";
  if (name.startsWith("duration-") || name.startsWith("ease-")) return "motion";
  if (name === "radius") return "dimension";
  if (name.startsWith("density-")) return "density";
  return "other";
}

function hslToToken(hslValue) {
  const raw = hslValue.trim().split(/\s+/);
  if (raw.length === 3) {
    const h = raw[0].replace(/%$/, "");
    const s = raw[1].replace(/%$/, "");
    const l = raw[2].replace(/%$/, "");
    return { $value: "hsl(" + h + ", " + s + "%, " + l + "%)", $type: "color" };
  }
  return { $value: hslValue, $type: "color" };
}

function buildGroup(vars, group) {
  const result = {};
  for (const [name, value] of Object.entries(vars)) {
    if (categorize(name) !== group) continue;
    const tokenName = name.replace(/-/g, ".");
    if (group === "color") {
      result[tokenName] = hslToToken(value);
    } else if (group === "motion") {
      result[tokenName] = {
        $value: value,
        $type: name.startsWith("duration") ? "duration" : "cubicBezier",
      };
    } else if (group === "dimension") {
      result[tokenName] = { $value: value, $type: "dimension" };
    } else if (group === "density") {
      result[tokenName] = { $value: value, $type: "dimension" };
    }
  }
  return result;
}

const tokens = {
  $schema: "https://design-tokens.github.io/community-group/format/",
  color: {
    light: buildGroup(lightVars, "color"),
    dark: buildGroup(darkVars, "color"),
  },
  motion: buildGroup(lightVars, "motion"),
  dimension: buildGroup(lightVars, "dimension"),
  density: {
    compact: buildGroup(extractBlock(css, '[data-density="compact"]'), "density"),
    comfortable: buildGroup(extractBlock(css, '[data-density="comfortable"]'), "density"),
    dense: buildGroup(extractBlock(css, '[data-density="dense"]'), "density"),
  },
};

writeFileSync(outputPath, JSON.stringify(tokens, null, 2) + "\n");
console.log(`Generated ${outputPath}`);
