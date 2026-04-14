import { SubstackClient } from "substack-api";
import { readFileSync } from "fs";
import { config } from "dotenv";

config();

// ── helpers ──────────────────────────────────────────────────

function usage(): never {
  console.error("Usage:");
  console.error(
    "  npm start -- me                            Test connection & show profile"
  );
  console.error(
    "  npm start -- draft <markdown-file>         Create article as draft"
  );
  console.error(
    "  npm start -- publish <markdown-file>       Publish article immediately"
  );
  console.error(
    "  npm start -- note <text>                   Publish a short note"
  );
  console.error(
    "  npm start -- list                          List your recent posts"
  );
  console.error("");
  console.error("Markdown file should have YAML frontmatter:");
  console.error("  ---");
  console.error("  title: My Article Title");
  console.error("  subtitle: Optional subtitle");
  console.error("  tags: go, buildinpublic");
  console.error("  ---");
  console.error("  Article body here (markdown, will be converted to HTML)...");
  process.exit(1);
}

interface Frontmatter {
  title: string;
  subtitle: string;
  tags: string[];
}

function parseFrontmatter(raw: string): { meta: Frontmatter; body: string } {
  const trimmed = raw.trim();
  if (!trimmed.startsWith("---")) {
    return { meta: { title: "Untitled", subtitle: "", tags: [] }, body: raw };
  }

  const end = trimmed.indexOf("---", 3);
  if (end === -1) {
    return { meta: { title: "Untitled", subtitle: "", tags: [] }, body: raw };
  }

  const yamlBlock = trimmed.slice(3, end).trim();
  const body = trimmed.slice(end + 3).trim();

  let title = "Untitled";
  let subtitle = "";
  let tags: string[] = [];

  for (const line of yamlBlock.split("\n")) {
    const colonIdx = line.indexOf(":");
    if (colonIdx === -1) continue;
    const key = line.slice(0, colonIdx).trim().toLowerCase();
    const val = line.slice(colonIdx + 1).trim();

    if (key === "title") title = val;
    else if (key === "subtitle") subtitle = val;
    else if (key === "tags")
      tags = val
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
  }

  return { meta: { title, subtitle, tags }, body };
}

/** Simple markdown → HTML for Substack's body field */
/** ProseMirror node types for Substack's editor */
interface PMNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: PMNode[];
  text?: string;
  marks?: { type: string; attrs?: Record<string, unknown> }[];
}

/** Convert markdown to Substack's ProseMirror JSON document format.
 *  Substack uses a ProseMirror-based editor that expects body_json as a
 *  stringified {type:"doc", content:[...]} structure.
 */
function markdownToProseMirror(md: string): string {
  const doc: PMNode = { type: "doc", content: [] };
  const lines = md.split("\n");
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Skip empty lines
    if (line.trim() === "") {
      i++;
      continue;
    }

    // Fenced code block
    const codeMatch = line.match(/^```(\w*)$/);
    if (codeMatch) {
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].match(/^```$/)) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // skip closing ```
      doc.content!.push({
        type: "codeBlock",
        attrs: { language: codeMatch[1] || null },
        content: [{ type: "text", text: codeLines.join("\n") }],
      });
      continue;
    }

    // Headings
    const h3Match = line.match(/^### (.+)$/);
    if (h3Match) {
      doc.content!.push({
        type: "heading",
        attrs: { level: 3 },
        content: parseInline(h3Match[1]),
      });
      i++;
      continue;
    }
    const h2Match = line.match(/^## (.+)$/);
    if (h2Match) {
      doc.content!.push({
        type: "heading",
        attrs: { level: 2 },
        content: parseInline(h2Match[1]),
      });
      i++;
      continue;
    }
    const h1Match = line.match(/^# (.+)$/);
    if (h1Match) {
      doc.content!.push({
        type: "heading",
        attrs: { level: 1 },
        content: parseInline(h1Match[1]),
      });
      i++;
      continue;
    }

    // Unordered list — collect consecutive "- " lines
    if (line.match(/^- /)) {
      const items: PMNode[] = [];
      while (i < lines.length && lines[i].match(/^- /)) {
        const itemText = lines[i].replace(/^- /, "");
        items.push({
          type: "listItem",
          content: [
            { type: "paragraph", content: parseInline(itemText) },
          ],
        });
        i++;
      }
      doc.content!.push({ type: "bulletList", content: items });
      continue;
    }

    // Ordered list — collect consecutive "1. " style lines
    if (line.match(/^\d+\. /)) {
      const items: PMNode[] = [];
      while (i < lines.length && lines[i].match(/^\d+\. /)) {
        const itemText = lines[i].replace(/^\d+\. /, "");
        items.push({
          type: "listItem",
          content: [
            { type: "paragraph", content: parseInline(itemText) },
          ],
        });
        i++;
      }
      doc.content!.push({ type: "orderedList", attrs: { start: 1 }, content: items });
      continue;
    }

    // Table (pipe-delimited)
    if (line.includes("|") && line.trim().startsWith("|")) {
      // Tables aren't natively supported in ProseMirror; convert to paragraphs
      // Skip separator row (|---|---|)
      const tableLines: string[] = [];
      while (i < lines.length && lines[i].includes("|") && lines[i].trim().startsWith("|")) {
        if (!lines[i].match(/^\|\s*-+/)) {
          tableLines.push(lines[i]);
        }
        i++;
      }
      for (const tl of tableLines) {
        const cells = tl.split("|").filter((c) => c.trim()).map((c) => c.trim());
        doc.content!.push({
          type: "paragraph",
          content: parseInline(cells.join(" | ")),
        });
      }
      continue;
    }

    // Regular paragraph
    doc.content!.push({
      type: "paragraph",
      content: parseInline(line),
    });
    i++;
  }

  return JSON.stringify(doc);
}

/** Parse inline markdown (bold, italic, code, links) into ProseMirror text nodes */
function parseInline(text: string): PMNode[] {
  const nodes: PMNode[] = [];
  // Regex to match inline elements: **bold**, *italic*, `code`, [text](url)
  const pattern =
    /(\*\*(.+?)\*\*)|(`([^`]+)`)|(\[([^\]]+)\]\(([^)]+)\))|((?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*))/g;
  let lastIndex = 0;
  let match;

  while ((match = pattern.exec(text)) !== null) {
    // Text before the match
    if (match.index > lastIndex) {
      nodes.push({ type: "text", text: text.slice(lastIndex, match.index) });
    }

    if (match[1]) {
      // **bold**
      nodes.push({
        type: "text",
        text: match[2],
        marks: [{ type: "bold" }],
      });
    } else if (match[3]) {
      // `code`
      nodes.push({
        type: "text",
        text: match[4],
        marks: [{ type: "code" }],
      });
    } else if (match[5]) {
      // [text](url)
      nodes.push({
        type: "text",
        text: match[6],
        marks: [{ type: "link", attrs: { href: match[7] } }],
      });
    } else if (match[8]) {
      // *italic*
      nodes.push({
        type: "text",
        text: match[9],
        marks: [{ type: "italic" }],
      });
    }

    lastIndex = match.index + match[0].length;
  }

  // Remaining text
  if (lastIndex < text.length) {
    nodes.push({ type: "text", text: text.slice(lastIndex) });
  }

  // If nothing was parsed, return at least one text node
  if (nodes.length === 0 && text) {
    nodes.push({ type: "text", text });
  }

  return nodes;
}



// ── env / config ─────────────────────────────────────────────

function getEnv(): { apiKey: string; hostname: string } {
  const apiKey = process.env.SUBSTACK_TOKEN;
  if (!apiKey) {
    console.error(
      "Error: SUBSTACK_TOKEN not set. Add your substack.sid cookie to .env"
    );
    console.error("See .env.example for instructions.");
    process.exit(1);
  }

  const hostname =
    process.env.SUBSTACK_PUBLICATION_URL || "substack.com";

  return { apiKey, hostname };
}

function getClient(): SubstackClient {
  const { apiKey, hostname } = getEnv();
  return new SubstackClient({ hostname, apiKey });
}

// ── direct Substack API for article creation ─────────────────
// The substack-api npm library only supports notes, not articles.
// Article drafts use Substack's internal REST API directly.

interface DraftResponse {
  id: number;
  title: string;
  slug: string;
  draft_title: string;
  type: string;
  [key: string]: unknown;
}

/**
 * Make authenticated requests to Substack.
 * account-level endpoints (e.g. /api/v1/subscription) → substack.com
 * publication endpoints (e.g. /api/v1/drafts)          → <publication>.substack.com
 */
async function substackFetch(
  path: string,
  method: string,
  body?: unknown,
  useGlobal = false
): Promise<unknown> {
  const { apiKey, hostname } = getEnv();
  const base = useGlobal ? "https://substack.com" : `https://${hostname}`;
  const url = `${base}${path}`;

  const res = await fetch(url, {
    method,
    headers: {
      Cookie: `connect.sid=${apiKey}`,
      "Content-Type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`HTTP ${res.status}: ${res.statusText} — ${text}`);
  }

  return res.json();
}

async function getAuthorId(): Promise<number> {
  const sub = (await substackFetch(
    "/api/v1/subscription",
    "GET"
  )) as { user_id: number; [key: string]: unknown };
  return sub.user_id;
}

async function createDraft(
  title: string,
  subtitle: string,
  bodyJson: string
): Promise<DraftResponse> {
  const authorId = await getAuthorId();

  // Step 1: create an empty draft with byline
  const draft = (await substackFetch("/api/v1/drafts", "POST", {
    draft_bylines: [{ id: authorId, is_guest: false }],
    type: "newsletter",
  })) as DraftResponse;

  // Step 2: update the draft with content (ProseMirror JSON format)
  const updated = (await substackFetch(
    `/api/v1/drafts/${draft.id}`,
    "PUT",
    {
      draft_title: title,
      draft_subtitle: subtitle,
      draft_body: bodyJson,
      draft_bylines: [{ id: authorId, is_guest: false }],
      type: "newsletter",
    }
  )) as DraftResponse;

  return updated;
}

async function publishDraft(draftId: number): Promise<unknown> {
  return substackFetch(`/api/v1/drafts/${draftId}/publish`, "POST", {
    send: false, // don't email subscribers, just publish on web
  });
}

// ── commands ─────────────────────────────────────────────────

async function cmdMe() {
  const client = getClient();
  console.log("Testing connectivity...");

  const ok = await client.testConnectivity();
  if (!ok) {
    console.error("Failed to connect — check your substack.sid cookie.");
    process.exit(1);
  }

  const profile = await client.ownProfile();
  console.log("\n✅ Connected to Substack");
  console.log(`   Name:   ${profile.name}`);
  console.log(`   Handle: @${profile.slug}`);
  console.log(`   Bio:    ${profile.bio || "(none)"}`);
}

async function cmdDraft(filePath: string) {
  const raw = readFileSync(filePath, "utf-8");
  const { meta, body } = parseFrontmatter(raw);
  const bodyJson = markdownToProseMirror(body);

  console.log(`Creating draft: "${meta.title}"...`);

  const draft = await createDraft(meta.title, meta.subtitle, bodyJson);

  const { hostname } = getEnv();
  console.log("\n✅ Draft created on Substack");
  console.log(`   Title:  ${meta.title}`);
  console.log(`   ID:     ${draft.id}`);
  console.log(`   Edit:   https://${hostname}/publish/post/${draft.id}`);
}

async function cmdPublish(filePath: string) {
  const raw = readFileSync(filePath, "utf-8");
  const { meta, body } = parseFrontmatter(raw);
  const bodyJson = markdownToProseMirror(body);

  console.log(`Publishing: "${meta.title}"...`);

  const draft = await createDraft(meta.title, meta.subtitle, bodyJson);
  await publishDraft(draft.id);

  const { hostname } = getEnv();
  const slug = draft.slug || draft.id;
  console.log("\n🚀 Published on Substack!");
  console.log(`   Title:  ${meta.title}`);
  console.log(`   URL:    https://${hostname}/p/${slug}`);
}

async function cmdNote(text: string) {
  const client = getClient();
  console.log("Publishing note...");

  const profile = await client.ownProfile();
  const note = await profile
    .newNote()
    .paragraph()
    .text(text)
    .publish();

  console.log("\n📝 Note published on Substack!");
  console.log(`   ID: ${(note as { id?: unknown }).id ?? "(created)"}`);
}

async function cmdList() {
  const client = getClient();
  const profile = await client.ownProfile();

  console.log(`\n📄 Recent posts by ${profile.name}:\n`);

  let count = 0;
  for await (const post of profile.posts({ limit: 10 })) {
    count++;
    const date = post.publishedAt?.toLocaleDateString() || "no date";
    console.log(`  ${count}. "${post.title}"`);
    console.log(`     Date:  ${date}`);
    console.log(`     Likes: ${post.likesCount}`);
    console.log("");
  }

  if (count === 0) {
    console.log("  No posts found.");
  }
}

// ── main ─────────────────────────────────────────────────────

async function main() {
  const args = process.argv.slice(2);
  if (args.length === 0) usage();

  const command = args[0].toLowerCase();

  try {
    switch (command) {
      case "me":
        await cmdMe();
        break;

      case "draft":
        if (!args[1]) {
          console.error("Error: provide a markdown file path");
          process.exit(1);
        }
        await cmdDraft(args[1]);
        break;

      case "publish":
        if (!args[1]) {
          console.error("Error: provide a markdown file path");
          process.exit(1);
        }
        await cmdPublish(args[1]);
        break;

      case "note":
        if (!args[1]) {
          console.error("Error: provide note text");
          process.exit(1);
        }
        await cmdNote(args.slice(1).join(" "));
        break;

      case "list":
        await cmdList();
        break;

      default:
        console.error(`Unknown command: ${command}`);
        usage();
    }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err);
    if (msg.includes("401")) {
      console.error(
        "\n❌ Authentication failed — your substack.sid cookie may have expired."
      );
      console.error(
        "   Re-extract it from your browser (F12 → Application → Cookies)."
      );
    } else if (msg.includes("403")) {
      console.error(
        "\n❌ Permission denied — check your account permissions."
      );
    } else if (msg.includes("429")) {
      console.error("\n❌ Rate limited — wait a minute and try again.");
    } else {
      console.error(`\n❌ Error: ${msg}`);
    }
    process.exit(1);
  }
}

main();
