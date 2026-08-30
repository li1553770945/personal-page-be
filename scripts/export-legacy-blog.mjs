import { createHash } from "node:crypto";
import { promises as fs } from "node:fs";
import path from "node:path";

const backendRoot = path.resolve(import.meta.dirname, "..");
const sourceRoot = path.resolve(backendRoot, "..", "blog", "docs");
const outputPath = path.join(backendRoot, "biz", "internal", "service", "blog", "legacy_posts.json");

const excludedRootFiles = new Set(["index.md", "about.md", "links.md"]);

async function walk(directory) {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    if (entry.name === ".vuepress" || entry.name === "@pages") continue;
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await walk(fullPath)));
    else if (entry.isFile() && entry.name.endsWith(".md")) files.push(fullPath);
  }
  return files;
}

function unquote(value) {
  const trimmed = value.trim();
  if (trimmed.length >= 2 && ((trimmed.startsWith('"') && trimmed.endsWith('"')) || (trimmed.startsWith("'") && trimmed.endsWith("'")))) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

function parseFrontmatter(source) {
  const normalized = source.replace(/^\uFEFF/, "").replace(/\r\n/g, "\n");
  if (!normalized.startsWith("---\n")) return { data: {}, content: normalized };
  const end = normalized.indexOf("\n---\n", 4);
  if (end < 0) return { data: {}, content: normalized };
  const header = normalized.slice(4, end).split("\n");
  const data = {};
  let currentArray = null;
  for (const line of header) {
    const item = line.match(/^\s*-\s+(.*)$/);
    if (item && currentArray) {
      data[currentArray].push(unquote(item[1]));
      continue;
    }
    const field = line.match(/^([A-Za-z][A-Za-z0-9_-]*):\s*(.*)$/);
    if (!field) continue;
    const [, key, rawValue] = field;
    if (rawValue.trim() === "" && (key === "categories" || key === "tags")) {
      data[key] = [];
      currentArray = key;
      continue;
    }
    currentArray = null;
    data[key] = unquote(rawValue);
  }
  return { data, content: normalized.slice(end + 5).trimStart() };
}

function descriptionFrom(content) {
  const paragraph = content
    .split(/\n\s*\n/)
    .map((value) => value.trim())
    .find((value) => value && !value.startsWith("#") && !value.startsWith("```") && !value.startsWith(":::") && !value.startsWith("<"));
  if (!paragraph) return "";
  return paragraph
    .replace(/!\[[^\]]*\]\([^)]*\)/g, "")
    .replace(/\[([^\]]+)\]\([^)]*\)/g, "$1")
    .replace(/[*_`>#]/g, "")
    .replace(/\s+/g, " ")
    .slice(0, 240);
}

const files = (await walk(sourceRoot)).filter((file) => {
  const relative = path.relative(sourceRoot, file).replace(/\\/g, "/");
  return relative.includes("/") || !excludedRootFiles.has(relative);
});

const posts = [];
for (const file of files) {
  const relativePath = path.relative(sourceRoot, file).replace(/\\/g, "/");
  const raw = await fs.readFile(file, "utf8");
  const { data, content: rawContent } = parseFrontmatter(raw);
  const content = rawContent.replaceAll("http://img.peacesheep.xyz", "https://img.peacesheep.xyz");
  const legacyPermalink = typeof data.permalink === "string" ? data.permalink : "";
  const permalinkMatch = legacyPermalink.match(/\/pages\/([^/]+)/);
  const slug = permalinkMatch?.[1] ?? createHash("sha256").update(relativePath).digest("hex").slice(0, 12);
  const fallbackTitle = path.basename(file, ".md").replace(/^\d+\.\s*/, "");
  posts.push({
    slug,
    legacyPermalink,
    title: String(data.title || fallbackTitle).trim(),
    description: String(data.description || descriptionFrom(content)).trim(),
    contentMarkdown: content,
    categories: Array.isArray(data.categories) ? data.categories.filter(Boolean) : [],
    tags: Array.isArray(data.tags) ? data.tags.filter(Boolean) : [],
    publishedAt: String(data.date || "").trim(),
    sourcePath: relativePath,
  });
}

posts.sort((a, b) => a.publishedAt.localeCompare(b.publishedAt));
const output = `${JSON.stringify(posts, null, 2)}\n`;
if (process.argv.includes("--stdout")) {
  process.stdout.write(output);
} else {
  await fs.writeFile(outputPath, output, "utf8");
  console.log(`Exported ${posts.length} legacy blog posts to ${outputPath}`);
}
