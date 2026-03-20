import { Hocuspocus } from "@hocuspocus/server";
import { Pool } from "pg";
import * as Y from "yjs";
import * as crypto from "crypto";

const pool = new Pool({
  connectionString:
    process.env.DATABASE_URL ??
    "postgres://postgres:postgres@localhost:5432/colign",
});

const JWT_SECRET = process.env.JWT_SECRET ?? "dev-secret-change-in-production";

function verifyJwt(token: string): { user_id: number; email: string; name: string } | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;

    // Verify signature
    const header = parts[0];
    const payload = parts[1];
    const signature = parts[2];
    const expected = crypto
      .createHmac("sha256", JWT_SECRET)
      .update(`${header}.${payload}`)
      .digest("base64url");

    if (signature !== expected) return null;

    const decoded = JSON.parse(Buffer.from(payload, "base64url").toString());

    // Check expiry
    if (decoded.exp && decoded.exp < Date.now() / 1000) return null;

    return decoded;
  } catch {
    return null;
  }
}

const server = new Hocuspocus({
  port: Number(process.env.PORT ?? 1234),

  async onAuthenticate({ token }) {
    const payload = verifyJwt(token);
    if (!payload) {
      throw new Error("Unauthorized");
    }
    return { user: payload };
  },

  async onLoadDocument({ documentName, document }) {
    // documentName format: "change-{id}-{docType}"
    const parts = documentName.split("-");
    if (parts.length < 3) return;

    const changeId = parts[1];
    const docType = parts.slice(2).join("-");

    try {
      const result = await pool.query(
        "SELECT content FROM documents WHERE change_id = $1 AND type = $2 LIMIT 1",
        [changeId, docType],
      );

      if (result.rows.length > 0 && result.rows[0].content) {
        const yXmlFragment = document.getXmlFragment("default");
        if (yXmlFragment.length === 0) {
          // Fix corrupted HTML with Tiptap node names
          const html = fixTiptapNodeNames(result.rows[0].content);
          const yMeta = document.getMap("meta");
          yMeta.set("initialHtml", html);
        }
      }
    } catch (err) {
      console.error("Failed to load document:", err);
    }
  },

  async onStoreDocument({ documentName, document }) {
    // Save Y.js document content back to DB
    const parts = documentName.split("-");
    if (parts.length < 3) return;

    const changeId = parts[1];
    const docType = parts.slice(2).join("-");

    try {
      // Get HTML from Y.js XML fragment
      const yXmlFragment = document.getXmlFragment("default");
      const content = yXmlFragmentToHtml(yXmlFragment);

      if (!content) return;

      // Upsert document (unique on change_id, type, title)
      await pool.query(
        `INSERT INTO documents (change_id, type, title, content, version)
         VALUES ($1, $2, '', $3, 1)
         ON CONFLICT (change_id, type, title)
         DO UPDATE SET content = $3, version = documents.version + 1, updated_at = NOW()`,
        [changeId, docType, content],
      );
    } catch (err) {
      console.error("Failed to store document:", err);
    }
  },

  async onConnect() {
    console.log("Client connected");
  },

  async onDisconnect() {
    console.log("Client disconnected");
  },
});

// Fix corrupted HTML that used Tiptap node names instead of HTML tags
function fixTiptapNodeNames(html: string): string {
  return html
    .replace(/<paragraph>/g, "<p>")
    .replace(/<\/paragraph>/g, "</p>")
    .replace(/<heading level="(\d)">/g, (_m, level) => `<h${level}>`)
    .replace(/<\/heading>/g, (match, offset, str) => {
      // Find the matching opening tag to get the level
      const before = str.substring(0, offset);
      const lastOpen = before.lastIndexOf("<h");
      if (lastOpen >= 0) {
        const level = before[lastOpen + 2];
        return `</h${level}>`;
      }
      return "</h2>";
    })
    .replace(/<bulletList>/g, "<ul>")
    .replace(/<\/bulletList>/g, "</ul>")
    .replace(/<orderedList>/g, "<ol>")
    .replace(/<\/orderedList>/g, "</ol>")
    .replace(/<listItem>/g, "<li>")
    .replace(/<\/listItem>/g, "</li>")
    .replace(/<codeBlock[^>]*>/g, "<pre><code>")
    .replace(/<\/codeBlock>/g, "</code></pre>")
    .replace(/<hardBreak\s*\/?>/g, "<br />")
    .replace(/<horizontalRule\s*\/?>/g, "<hr />");
}

// Map Tiptap node names to HTML tags
const NODE_TO_TAG: Record<string, string> = {
  paragraph: "p",
  bulletList: "ul",
  orderedList: "ol",
  listItem: "li",
  blockquote: "blockquote",
  codeBlock: "pre",
  hardBreak: "br",
  horizontalRule: "hr",
  taskList: "ul",
  taskItem: "li",
};

const SELF_CLOSING = new Set(["br", "hr", "img"]);

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

function yXmlFragmentToHtml(fragment: Y.XmlFragment): string {
  let html = "";
  fragment.forEach((item) => {
    if (item instanceof Y.XmlElement) {
      html += xmlElementToHtml(item);
    } else if (item instanceof Y.XmlText) {
      html += xmlTextToHtml(item);
    }
  });
  return html;
}

function xmlTextToHtml(text: Y.XmlText): string {
  let result = "";
  const delta = text.toDelta();
  for (const op of delta) {
    if (typeof op.insert === "string") {
      let content = escapeHtml(op.insert);
      if (op.attributes) {
        if (op.attributes.bold) content = `<strong>${content}</strong>`;
        if (op.attributes.italic) content = `<em>${content}</em>`;
        if (op.attributes.code) content = `<code>${content}</code>`;
        if (op.attributes.underline) content = `<u>${content}</u>`;
        if (op.attributes.strike) content = `<s>${content}</s>`;
        if (op.attributes.commentHighlight) {
          const commentId = op.attributes.commentHighlight.commentId;
          if (commentId) {
            content = `<span data-comment-id="${commentId}" class="comment-highlight">${content}</span>`;
          }
        }
      }
      result += content;
    }
  }
  return result;
}

function xmlElementToHtml(element: Y.XmlElement): string {
  const nodeName = element.nodeName;
  const attrs = element.getAttributes();

  // Resolve tag name
  let tag: string;
  if (nodeName === "heading") {
    const level = attrs.level || 1;
    tag = `h${level}`;
  } else {
    tag = NODE_TO_TAG[nodeName] || nodeName;
  }

  // Build attribute string (skip Tiptap-internal attrs)
  let attrStr = "";
  for (const [key, value] of Object.entries(attrs)) {
    if (nodeName === "heading" && key === "level") continue;
    if (nodeName === "taskItem" && key === "checked") {
      attrStr += ` data-checked="${value}"`;
      continue;
    }
    attrStr += ` ${key}="${value}"`;
  }
  if (nodeName === "taskList") attrStr += ' data-type="taskList"';
  if (nodeName === "taskItem") attrStr += ' data-type="taskItem"';

  // Self-closing tags
  if (SELF_CLOSING.has(tag)) {
    return `<${tag}${attrStr} />`;
  }

  // Build inner content
  let inner = "";
  element.forEach((child) => {
    if (child instanceof Y.XmlElement) {
      inner += xmlElementToHtml(child);
    } else if (child instanceof Y.XmlText) {
      inner += xmlTextToHtml(child);
    }
  });

  // codeBlock wraps content in <code>
  if (nodeName === "codeBlock") {
    inner = `<code>${inner}</code>`;
  }

  return `<${tag}${attrStr}>${inner}</${tag}>`;
}

server.listen();
console.log(`Hocuspocus listening on port ${process.env.PORT ?? 1234}`);
