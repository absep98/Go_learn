---
name: draft-post
description: "Draft a build-in-public post based on recent work in this repo. Use when: writing a LinkedIn post, dev.to article, Twitter/X thread, Substack post, or any public post about what you have built or learned. Reads git history and codebase to generate ready-to-copy content."
argument-hint: "platform and topic — e.g. 'LinkedIn post about JWT auth' or 'dev.to article about rate limiting'"
---

# Draft Build-in-Public Post

Generates ready-to-copy social/blog posts based on what has actually been built in this Go learning + backend project.

## When to Use
- Writing a LinkedIn update about recent progress
- Drafting a dev.to / Hashnode article about a feature or concept
- Composing a Twitter/X thread about something you learned
- Writing a Substack or any other platform post
- Generating a weekly progress summary

## Procedure

### Step 0 — Check for Auto-Captured Commit Context
**Always do this first.** A git post-commit hook automatically saves the latest commit info to:
```
.github/skills/draft-post/scripts/.last-commit-context.json
```
Read this file. It contains `hash`, `message`, `files_changed`, `stat`, and `date` for the most recent commit. Use this as the primary input for the post — the user just committed something and wants to write about it.

If the file doesn't exist or the user specifies a different topic, fall back to Step 1.

### Step 1 — Gather Additional Context
Read the changed files from the commit context to understand what was actually built:
```powershell
cd "c:\learning\Go_learn"
git show <hash> --stat
git diff <hash>~1 <hash>
```

Then read relevant source files based on the topic. Key areas:
- `Project-Building-While-Learning/personal-analytics-backend/internal/handlers/` — auth, middleware, entries, webhooks
- `Project-Building-While-Learning/personal-analytics-backend/internal/db/` — database layer
- `Project-Building-While-Learning/personal-analytics-backend/ARCHITECTURE-MAP.md` — overall system design
- `Project-Building-While-Learning/personal-analytics-backend/LEARNING-JOURNEY.md` — personal learning notes

### Step 2 — Identify the Story
Extract a clear angle from the context:
- A specific technical problem solved (e.g., "how I added rate limiting")
- A mistake made and fixed
- A concept explained simply (e.g., "what middleware actually does")
- A milestone reached (e.g., "my first Go REST API is live")

### Step 3 — Apply the Right Template

Use the template from [./references/templates.md](./references/templates.md) that matches the target platform.

### Step 4 — Output & Post
- Provide the full post ready to copy-paste
- Save the article as a `.md` file in `.github/skills/draft-post/scripts/`
- Note the character/word count
- Suggest 3-5 relevant hashtags
- Suggest one follow-up post idea
- Ask: "Ready to post? I'll push to all platforms as draft (you can review first) or publish directly."
- Run `post-everywhere.ps1` with the user's chosen mode

## Platform Guidelines

| Platform | Length | Tone | Format |
|---|---|---|---|
| LinkedIn | 150-300 words | Professional but personal | Short paragraphs, no headers |
| Twitter/X thread | 5-10 tweets, 280 chars each | Casual, punchy | Numbered tweets |
| dev.to / Hashnode | 500-1500 words | Technical + conversational | Headers, code blocks |
| Substack | 300-800 words | Personal essay style | Flowing prose, one key insight |

## Your Voice Guidelines
- You are learning Go in public — own the journey, including struggles
- Be specific: name the exact thing (JWT, Redis, bcrypt, goroutines) not vague ("I built something cool")
- Show before/after or problem/solution structure when possible
- Always include a GitHub link: https://github.com/absep98/Go_learn

## Post Everywhere (Unified — Recommended)

One command posts the same article to **dev.to + Hashnode + Substack** at once:

```powershell
cd .github/skills/draft-post/scripts

# Draft on all platforms (safe — nothing published)
.\post-everywhere.ps1 -File my-article.md -Mode draft

# Publish on all platforms
.\post-everywhere.ps1 -File my-article.md -Mode publish
```

The script automatically skips any platform that doesn't have a `.env` configured.

### Full Skill Workflow

When the user invokes `/draft-post`:
1. **Generate** the markdown article following the Procedure above → save to a `.md` file in the scripts folder
2. **Show** the article to the user for review/edits
3. **Ask** the user: "Draft or publish? Which platforms?" (default: all, draft)
4. **Run** `post-everywhere.ps1` with the chosen mode
5. **Report** the results table showing status per platform

### Article Markdown Format (shared across all platforms)
```markdown
---
title: My Article Title
subtitle: Optional subtitle (Substack only)
tags: go, buildinpublic, backend
series: Learning Go in Public
---
Article body here in markdown...
```

## Individual Platform Scripts

### Posting to X (Twitter)

After drafting a tweet or thread, you can post it directly using the Go script:

```powershell
cd .github/skills/draft-post/scripts/post-to-x
go run main.go "Your tweet text here"
```

Or pipe from a file:
```powershell
echo "Your tweet text" | go run main.go
```

**Setup:** Copy `.env.example` to `.env` and fill in your X API credentials (Consumer Key, Consumer Secret, Access Token, Access Token Secret). See [post-to-x script](./scripts/post-to-x/main.go).

> **Note:** X API requires Pay-per-Use credits. If you're on the deprecated Free tier, the API returns 503.

## Posting to dev.to (FREE)

Post articles directly from markdown files with frontmatter:

```powershell
cd .github/skills/draft-post/scripts/post-to-devto

# Save as draft first (safe — not published)
go run main.go draft my-article.md

# Publish when ready
go run main.go publish my-article.md

# List your articles
go run main.go list
```

Markdown file format:
```markdown
---
title: My Article Title
tags: go, buildinpublic, backend
series: Learning Go in Public
---
Article body here...
```

**Setup:** Copy `.env.example` to `.env` and add your dev.to API key from https://dev.to/settings/extensions. See [post-to-devto script](./scripts/post-to-devto/main.go).

## Posting to Hashnode (FREE)

Post articles to your Hashnode blog via GraphQL API:

```powershell
cd .github/skills/draft-post/scripts/post-to-hashnode

# First time: find your publication ID
go run main.go me

# Save as draft
go run main.go draft my-article.md

# Publish directly
go run main.go publish my-article.md

# List your posts
go run main.go list
```

Uses the same markdown frontmatter format as dev.to.

**Setup:** Copy `.env.example` to `.env`, add your token from https://hashnode.com/settings/developer, run `go run main.go me` to get your publication ID. See [post-to-hashnode script](./scripts/post-to-hashnode/main.go).

## Posting to Substack (FREE — cookie auth)

Post articles and notes to your Substack newsletter:

```powershell
cd .github/skills/draft-post/scripts/post-to-substack

# First time: test your cookie and see profile
npm start -- me

# Save article as draft (editable in Substack dashboard)
npm start -- draft my-article.md

# Publish article immediately (web only, no email)
npm start -- publish my-article.md

# Publish a short note (like a tweet)
npm start -- note "🚀 Just shipped a new feature! Learning Go in public."

# List your recent posts
npm start -- list
```

Uses the same markdown frontmatter format as dev.to/Hashnode, plus optional `subtitle:` field.

**Setup:**
1. Copy `.env.example` to `.env`
2. Login to Substack in your browser
3. Open DevTools (F12) → Application tab → Cookies → `https://substack.com`
4. Copy the `substack.sid` cookie value into `SUBSTACK_TOKEN`
5. Set `SUBSTACK_PUBLICATION_URL` to your subdomain (e.g. `yourname.substack.com`)
6. Run `npm start -- me` to verify

> **Note:** Cookie auth expires periodically. If you get 401 errors, re-extract the cookie from your browser. See [post-to-substack script](./scripts/post-to-substack/src/main.ts).
