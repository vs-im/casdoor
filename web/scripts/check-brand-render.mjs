// Proves that the brand-free locale bundles render with a deployment's brand:
// every "{{brand}}" placeholder is substituted by i18next exactly the way
// src/i18n.ts configures it (interpolation.defaultVariables), with no call site
// passing anything.
//
//   node web/scripts/check-brand-render.mjs "Acme Identity"
//
// See docs/branding.md.
import fs from "fs";
import path from "path";
import {fileURLToPath} from "url";
import i18n from "i18next";

const brand = process.argv[2] || process.env.CASDOOR_BRAND_NAME || "Casdoor";
const localesDir = path.join(path.dirname(fileURLToPath(import.meta.url)), "..", "src", "locales");

const languages = fs.readdirSync(localesDir);
const resources = Object.fromEntries(languages.map((lang) => [
  lang,
  JSON.parse(fs.readFileSync(path.join(localesDir, lang, "data.json"), "utf8")),
]));

await i18n.init({
  lng: "en",
  resources,
  fallbackLng: "en",
  keySeparator: false,
  nsSeparator: ":",
  ns: Object.keys(resources.en),
  interpolation: {escapeValue: true, defaultVariables: {brand}},
});

let placeholders = 0;
let rendered = 0;
const failures = [];

for (const lang of languages) {
  await i18n.changeLanguage(lang);
  for (const [ns, entries] of Object.entries(resources[lang])) {
    for (const [key, value] of Object.entries(entries)) {
      if (typeof value !== "string" || !value.includes("{{brand}}")) {
        continue;
      }
      placeholders++;
      const text = i18n.t(`${ns}:${key}`);
      if (text.includes("{{brand}}") || !text.includes(brand)) {
        failures.push(`${lang}/${ns}:${key} -> ${text}`);
      } else {
        rendered++;
      }
    }
  }
}

if (placeholders === 0) {
  console.error("check-brand-render: no {{brand}} placeholder found in the locales, the branding hook is gone");
  process.exit(1);
}

if (failures.length > 0) {
  console.error(`check-brand-render: ${failures.length} strings did not render the brand:`);
  failures.slice(0, 10).forEach((line) => console.error(`  ${line}`));
  process.exit(1);
}

console.log(`check-brand-render: OK, ${rendered} strings in ${languages.length} languages render as "${brand}"`);
