# ⚡ VS Code 1.109 - Quick Reference Card

> Keep this open as a reference while you work!

---

## 🎯 Most Important Commands

| Command | Action |
|---------|--------|
| `/plan` | Create implementation plan before coding |
| `/init` | Generate workspace instructions for AI |
| `Ctrl+I` | Open inline chat (quick AI help) |
| `Ctrl+Shift+P` | Command Palette (all commands) |

---

## 🤖 Agent Types Quick Pick

```
┌──────────────┬─────────────────┬─────────────────────┐
│ Agent Type   │ When to Use     │ Access              │
├──────────────┼─────────────────┼─────────────────────┤
│ LOCAL        │ Quick questions │ Default in chat     │
│              │ Small edits     │                     │
│              │ Explanations    │                     │
├──────────────┼─────────────────┼─────────────────────┤
│ BACKGROUND   │ Parallel work   │ Session type picker │
│              │ Refactoring     │ in chat input       │
│              │ While you code  │                     │
├──────────────┼─────────────────┼─────────────────────┤
│ CLOUD        │ Complex tasks   │ Session type picker │
│              │ Creates PRs     │ in chat input       │
│              │ Multi-file work │                     │
├──────────────┼─────────────────┼─────────────────────┤
│ CLAUDE       │ Advanced logic  │ Session type picker │
│              │ Reasoning tasks │ (Preview)           │
└──────────────┴─────────────────┴─────────────────────┘
```

---

## 🖱️ Mouse Actions (NEW!)

### Double-Click Selection

| Position | Result |
|----------|--------|
| After `{` | Selects content inside `{}` |
| Before `}` | Selects content inside `{}` |
| After `"` | Selects string content |
| Before `"` | Selects string content |
| After `(` | Selects content inside `()` |
| Before `)` | Selects content inside `()` |

**Example:** In `func main() { code here }`, double-click after `{` to select `code here`

---

## ⌨️ Keyboard Shortcuts

### Chat & AI

| Shortcut | Action |
|----------|--------|
| `Ctrl+I` | Inline chat |
| `Ctrl+Shift+I` | Open Chat view |
| `Esc` | Close inline chat |
| `1-9` | Select numbered answer in AI questions |
| `↑↓` | Navigate AI answer options |
| `Escape` | Skip remaining AI questions |

### Editor

| Shortcut | Action |
|----------|--------|
| `Alt+Click` | Multiple cursors |
| `Ctrl+D` | Select next occurrence |
| `Ctrl+Shift+L` | Select all occurrences |
| `Ctrl+/` | Toggle comment |
| `Alt+↑↓` | Move line up/down |

### Terminal

| Shortcut | Action |
|----------|--------|
| `` Ctrl+` `` | Toggle terminal |
| `Ctrl+Shift+5` | Split terminal |
| `Ctrl+Shift+C` | Copy from terminal |
| `Ctrl+Shift+V` | Paste to terminal |

---

## 📊 Chat Symbols (Context References)

```
#file           → Reference a specific file
#codebase       → Search entire codebase
#terminalLastCommand → Last terminal command
#terminalSelection  → Selected terminal text
```

**Example:** `#file:main.go explain the main function`

---

## 🔍 Problems Panel Filters

```
source:gopls          → Only Go language server issues
!source:gopls         → Everything except gopls
source:test           → Only test failures
error                 → Only errors (not warnings)
warning               → Only warnings
```

---

## 📤 Output Panel Filters

```
!debug                → Hide lines containing "debug"
error,warning         → Show lines with "error" OR "warning"
!test                 → Exclude lines with "test"
```

---

## 💬 Chat Slash Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `/plan` | Create implementation plan | `/plan build a CLI tool` |
| `/init` | Generate workspace instructions | `/init` |
| `/fix` | Fix selected code | `/fix this function` |
| `/explain` | Explain selected code | `/explain` |
| `/tests` | Generate tests | `/tests for this function` |
| `/doc` | Generate documentation | `/doc` |

---

## 🎨 Bracket Matching Colors

**Settings:**

```json
"workbench.colorCustomizations": {
  "editorBracketMatch.foreground": "#4EC9B0",  // Text color
  "editorBracketMatch.background": "#3a3d41",  // Background
  "editorBracketMatch.border": "#4EC9B0"       // Border
}
```

---

## 🖥️ Terminal Commands (Auto-Approved)

These run **without confirmation** when auto-approval is enabled:

```bash
cd, ls, dir          # Navigation
pwd, echo            # Info
go run, go test      # Go commands
go build, go mod     # Build commands
npm install, yarn    # Package managers
docker ps, docker images  # Safe docker
git status, git log  # Safe git
```

---

## 🔐 Terminal Sandboxing (Mac/Linux)

**Enable:**
```json
"chat.tools.terminal.sandbox.enabled": true
```

**Configure Network:**
```json
"chat.tools.terminal.sandbox.network": {
  "allowed": [
    "github.com",
    "golang.org",
    "localhost"
  ]
}
```

---

## 🌐 Integrated Browser

**Open:** `Browser: Open Integrated Browser`

**Features:**
- Full DevTools (F12)
- Authentication support
- Add element to chat
- Find in page (Ctrl+F)

**Settings:**
```json
"workbench.browser.openLocalhostLinks": true  // Auto-open localhost
```

---

## 📈 Context Window Tips

### Hover Indicator Shows:

- **Messages**: Your chat history
- **Tools**: Tool call results
- **Attachments**: Files you've referenced
- **Total**: Total tokens used / available

### When Near Limit:

1. Start new chat session
2. Remove old tool results
3. Unpin unnecessary files
4. Use subagents (they have separate context)

---

## 🛠️ Useful VS Code Commands

**Access via:** `Ctrl+Shift+P`

```
Chat: Configure Skills
Chat: New Skill File
Browser: Open Integrated Browser
Git: Delete
View: Toggle Agent Sessions
Developer: Reload Window (when things act weird)
```

---

## 📁 File Locations Reference

```
.vscode/
  └── settings.json                  → Project settings

copilot-instructions.md              → Main AI instructions
AGENTS.md                            → Alternative name

.github/
  ├── skills/                        → Agent skills
  └── agents/                        → Custom agent definitions

.claude/
  ├── skills/                        → Claude-specific skills
  └── agents/                        → Claude agent definitions
```

---

## 🎯 Quick Workflows

### Workflow 1: New Feature

```
1. /plan "feature description"
2. Review plan, ask questions
3. Switch to Background/Cloud agent
4. Implement while you continue work
5. Test in integrated browser
```

### Workflow 2: Debug Complex Code

```
1. Select problematic code
2. Inline chat: "explain this"
3. Ask for Mermaid diagram
4. Ask: "what could go wrong?"
5. Ask: "suggest fixes"
```

### Workflow 3: Learning New Concept

```
1. Ask: "explain [concept] in Go"
2. Ask: "show me a simple example"
3. Ask: "show me a real-world use case"
4. Ask: "generate practice exercises"
5. Ask: "create tests for my solution"
```

### Workflow 4: Code Review

```
1. Select your code
2. Ask: "review for Go best practices"
3. Ask: "check for potential bugs"
4. Ask: "suggest performance improvements"
5. Apply suggested changes
```

---

## 🚨 Troubleshooting Quick Fixes

| Problem | Solution |
|---------|----------|
| AI not using instructions | Check Chat Diagnostics (right-click → Diagnostics) |
| Settings not applying | Reload Window (`Ctrl+Shift+P` → reload) |
| Context window full | Start new chat or remove attachments |
| Terminal commands fail | Check sandbox settings (may be too restrictive) |
| Chat feels slow | Check context window usage, start fresh chat |
| Custom agent not appearing | Verify file in `.github/agents/` with `.agent.md` |

---

## 💾 Settings Cheat Sheet

### Most Important Settings

```json
{
  // AI & Chat
  "chat.thinking.style": "detailed",
  "github.copilot.chat.copilotMemory.enabled": true,
  "chat.useAgentSkills": true,
  
  // Terminal
  "chat.tools.terminal.enableAutoApprove": true,
  "terminal.integrated.enableKittyKeyboardProtocol": true,
  
  // Editor
  "editor.inlineSuggest.enabled": true,
  "editor.formatOnSave": true,
  
  // Browser
  "workbench.browser.openLocalhostLinks": true,
  
  // Go-specific
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true
  }
}
```

---

## 📱 Status Indicators

**Command Center (Top):**

- 💬 Chat button → Click to toggle chat view
- 🤖 Agent status → Shows active agents
- 🔴 → Agent needs attention
- 🟡 → Agent in progress
- 🔵 → Unread messages

**Chat Input (Bottom):**

- 📊 Context window → Hover for details
- 🎯 Session type → Click to switch agent
- 📎 Attachments → Files in context

---

## 🎓 Learn More Resources

| Resource | Link/Location |
|----------|---------------|
| Full guide | `VS_CODE_PRODUCTIVITY_GUIDE.md` |
| Quick start | `QUICK_START_CHECKLIST.md` |
| Your settings | `.vscode/settings.json` |
| AI instructions | `copilot-instructions.md` |
| Official docs | code.visualstudio.com/updates |

---

## 🏆 Power User Checklist

Track your mastery:

**Basics:**
- [ ] Used `/plan` for a project
- [ ] Tried inline chat 10+ times
- [ ] Double-click selection feels natural
- [ ] Understand context window

**Intermediate:**
- [ ] Used background agent
- [ ] Created workspace instructions
- [ ] Filtered problems by source
- [ ] Used integrated browser

**Advanced:**
- [ ] Created custom skill
- [ ] Used 3 agents in parallel
- [ ] Built Copilot Memory
- [ ] Generated Mermaid diagrams
- [ ] Configured terminal sandboxing

---

## 💡 Pro Tips

1. **Start conversations with context**: Instead of "explain this", say "explain this goroutine pattern in the context of my worker pool"

2. **Use progressive refinement**: Ask broad question first, then refine with follow-ups

3. **Leverage subagents**: For complex tasks, agent will use subagents automatically (they don't consume your context!)

4. **Name your patterns**: Tell AI "I call this pattern X" - it helps with memory

5. **Review thinking tokens**: When enabled, see why AI chose specific approaches

---

**Print this and keep it near your desk! 📌**

*Last updated: February 5, 2026 - VS Code 1.109*
