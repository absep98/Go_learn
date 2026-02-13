# VS Code 1.109 Productivity Guide for Go Developers

## 🎯 Daily Workflow Enhancements

### Starting Your Day

1. **Agent Sessions Welcome** (Experimental)
   - See recent AI sessions on startup
   - Quick resume of ongoing work
   - Enable: `"workbench.startupEditor": "agentSessionsWelcomePage"`

2. **Clean Editor Start**
   - Control whether previous tabs restore
   - Setting: `workbench.editor.restoreEditors`

---

## ⚡ Essential Keyboard Shortcuts & Actions

### Chat & AI

| Action | Command/Shortcut |
|--------|------------------|
| Open Chat | Check chat button in command center |
| Inline Chat | Select code → affordance appears |
| Plan Agent | Type `/plan` in chat |
| Init Workspace | Type `/init` in chat |
| New Local Chat | Cmd/Ctrl+Shift+P → "New Local Chat" |

### Editor Selection

| Action | How To |
|--------|--------|
| Select bracket content | Double-click after `{` or before `}` |
| Select string content | Double-click after `"` or before `"` |
| Select word | Double-click on word (existing) |

### Terminal

| Feature | Usage |
|---------|-------|
| Interactive input | Click in embedded terminal, type directly |
| Delete hidden terminals | Click delete icon on "Hidden Terminals" |
| View command details | Hover over terminal command in chat |

---

## 🤖 Working with Multiple Agents

### Agent Types & When to Use

```
┌─────────────────────────────────────────────┐
│ LOCAL AGENT                                 │
│ • Quick questions                           │
│ • Code explanations                         │
│ • Small edits                               │
│ Use: Default for most tasks                │
├─────────────────────────────────────────────┤
│ BACKGROUND AGENT                            │
│ • Refactoring while you work               │
│ • Test generation                           │
│ • Documentation updates                     │
│ Use: Parallel work, don't block yourself   │
├─────────────────────────────────────────────┤
│ CLOUD AGENT                                 │
│ • Complex feature implementation           │
│ • Large-scale refactoring                  │
│ • Multi-file changes                        │
│ Use: Heavy lifting, creates PR             │
├─────────────────────────────────────────────┤
│ CLAUDE AGENT (Preview)                     │
│ • Uses Anthropic's Claude SDK              │
│ • Advanced reasoning capabilities          │
│ Use: Complex problem-solving               │
└─────────────────────────────────────────────┘
```

### Switching Between Agents

1. Click **session type picker** in chat input
2. Select agent type
3. Optionally hand off current session

---

## 📊 Understanding Chat Context

### Context Window Indicator

**Location:** Chat input area (bottom)

**What to check:**
- Hover to see token breakdown
- Categories: messages, tools, attachments
- When near limit: start new session or remove context

### Best Practices

```
✅ DO:
- Start new chat for new topics
- Use /plan for complex tasks
- Attach only relevant files
- Clear old sessions regularly

❌ DON'T:
- Keep adding to infinite chat
- Attach entire codebase
- Mix unrelated questions
- Ignore context warnings
```

---

## 🔧 Go-Specific Productivity Tips

### 1. Testing Workflow

```go
// Use integrated browser for testing web servers
// Enable: workbench.browser.openLocalhostLinks

// In your Go code:
func main() {
    http.ListenAndServe(":8080", handler)
}

// VS Code will open localhost:8080 in integrated browser
// Features: DevTools, inspect elements, add to chat for help
```

### 2. Error Filtering

```
Problems Panel Filters:
  source:gopls          → Only Go language server
  !source:gopls         → Everything except gopls
  source:test           → Only test failures
```

### 3. Git Workflow

```
New Commands:
- Git: Delete → Safely remove file with git rm
- Collapse All → In SCM tree view
- Inline blame → Hover to see commit info
```

### 4. Terminal Commands

```bash
# Auto-approved safe commands (when enabled):
cd path/to/dir
ls, dir
docker ps
npm install
go mod download
go test ./...
```

---

## 🎨 Visual Enhancements

### Bracket Matching

```json
// Customize in settings.json:
"workbench.colorCustomizations": {
  "editorBracketMatch.foreground": "#your-color",
  "editorBracketMatch.background": "#your-color",
  "editorBracketMatch.border": "#your-color"
}
```

### Ghost Text (Inline Suggestions)

- Now shows **dotted underline** for short suggestions
- Makes it easier to see single-character suggestions like `)`

---

## 📁 Project Organization

### Workspace Instructions

**File:** `copilot-instructions.md` or `AGENTS.md`

**Generate:** Type `/init` in chat

**Example content:**
```markdown
# Go Project Conventions

## Code Style
- Use table-driven tests
- Error handling: return early
- Package naming: lowercase, no underscores

## Project Structure
- cmd/ for entry points
- internal/ for private packages
- Use interfaces for testability

## Testing
- Test file naming: *_test.go
- Benchmark naming: Benchmark*
- Use testify/assert for assertions
```

### Agent Skills

**Location:** `.github/skills/` or `.claude/skills/`

**Create:** Command Palette → "Chat: New Skill File"

**Example Skills:**
- `testing-strategy/` - How to write tests
- `error-handling/` - Go error patterns
- `documentation/` - Godoc conventions

---

## 🔐 Security Features

### Terminal Sandboxing (macOS/Linux)

```json
// Restrict agent terminal commands
"chat.tools.terminal.sandbox.enabled": true,
"chat.tools.terminal.sandbox.network": {
  "allowed": [
    "github.com",
    "golang.org",
    "localhost"
  ]
}
```

### Auto-Approval Rules

```json
// Safe commands auto-approved
"chat.tools.terminal.enableAutoApprove": true
```

---

## 🎓 Learning with AI Features

### For Go Learning Specifically

1. **Use Plan Agent for projects**
   ```
   /plan Create a CLI tool that parses JSON logs
   ```

2. **Ask for explanations**
   - Select code → Ask "explain this concurrency pattern"
   - AI shows thinking process (if enabled)

3. **Generate test cases**
   ```
   Select function → Ask "generate table-driven tests"
   ```

4. **Code reviews**
   ```
   Ask: "Review this code for Go best practices"
   ```

5. **Mermaid diagrams**
   ```
   Ask: "Show me a sequence diagram of this HTTP handler flow"
   ```

---

## 📈 Tracking Your Progress

### Agent Sessions View

**Access:** Sidebar → Agent Sessions

**Features:**
- See all active sessions
- Filter by status (in-progress, needs attention)
- Multi-select for bulk operations
- Resume any session

### Status Indicator

**Location:** Command center (top)

**Shows:**
- Active agents
- Unread messages
- Sessions needing attention

**Click:** Opens filtered session list

---

## 🛠️ Troubleshooting

### Chat Customization Diagnostics

**Access:** Right-click in Chat view → Diagnostics

**Shows:**
- Loaded agents, prompts, instructions, skills
- Load status and errors
- Source locations

**Use when:**
- Custom agents not working
- Instructions not applying
- Skills not being used

---

## 💡 Pro Tips

1. **Keyboard Navigation**
   - Use number keys to select answers in AI questions
   - Up/Down to navigate options
   - Escape to skip remaining questions

2. **Parallel Work**
   - Start background agent for refactoring
   - Continue coding in local agent
   - Check progress in Agent Sessions view

3. **Context Management**
   - Use subagents for isolated tasks (they don't use main context)
   - Enable search subagent for complex queries
   - Start new sessions for new topics

4. **Model Selection**
   - Hover over model in picker to see capabilities
   - Choose appropriate model for task
   - Configure defaults in settings

5. **Memory Usage**
   - Add memories for recurring preferences
   - Example: "When writing Go tests, use testify/assert"
   - Manages at GitHub Copilot settings

---

## 🎬 Getting Started Checklist

- [ ] Review `.vscode/settings.json` I created
- [ ] Enable Copilot Memory
- [ ] Try `/init` to generate workspace instructions
- [ ] Test double-click bracket selection
- [ ] Open Agent Sessions view
- [ ] Try `/plan` for your next feature
- [ ] Enable integrated browser for testing
- [ ] Set up terminal auto-approval
- [ ] Create a custom skill for Go best practices
- [ ] Try inline chat on selected code

---

## 📚 Quick Reference Commands

```
Chat Commands:
/plan              → Create implementation plan
/init              → Generate workspace instructions

VS Code Commands:
Browser: Open Integrated Browser
Chat: Configure Skills
Chat: New Skill File
Git: Delete
Announce Cursor Position (screen readers)

Settings to Explore:
chat.thinking.style
inlineChat.affordance
github.copilot.chat.copilot Memory.enabled
workbench.browser.openLocalhostLinks
chat.useAgentSkills
```

---

## 🌟 Next Steps

1. **Experiment** with different agent types for different tasks
2. **Create** workspace instructions for your Go projects
3. **Build** custom skills for recurring patterns
4. **Use** the Plan agent before starting complex features
5. **Monitor** context window usage
6. **Leverage** integrated browser for testing

---

*Last Updated: February 5, 2026*
*VS Code Version: 1.109 (January 2026)*
