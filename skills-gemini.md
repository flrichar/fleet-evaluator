Since you're working with Go and high-level infra (Kubernetes, NATS), you'll appreciate that the Gemini CLI treats "Skills" more like a **packaged module** than a simple text file. 

In the current Gemini CLI ecosystem (v2026), `SKILL.md` is the standard naming convention for local agent instructions. Here is how you can set this up to mirror your "Skills" workflow in the terminal.

---

## 1. The Skill Structure
Gemini CLI expects a specific directory structure to "auto-discover" your skills. Instead of just a loose `.md` file, wrap it like this:

```bash
# Recommended structure
my-project/
├── .agents/
│   └── skills/
│       └── go-expert/
│           ├── SKILL.md      # Your primary instructions
│           └── snippets.go   # (Optional) Reference code
```

## 2. Adapting your `SKILL.md`
Unlike Claude’s skills, Gemini requires a **YAML Frontmatter** block at the top of the file for the CLI to parse the metadata.

**Example `SKILL.md` for your Go/NATS context:**
```markdown
---
name: go-infra-expert
description: Expert in Golang, NATS messaging, and Kubernetes GitOps patterns.
---

# Instructions
You are a senior backend engineer. When I ask for code:
* Use idiomatic Go (Golang).
* Prefer NATS JetStream patterns for messaging.
* Ensure all K8s manifests follow GitOps (ArgoCD/Flux) best practices.

# Knowledge
- Internal NATS cluster address: nats://nats.infra.svc.cluster.local:4222
- Always use structured logging with 'slog'.
```

---

## 3. Implementation via CLI
Once your file is ready, you can "link" it so it's available globally or per-session.

### Linking the Skill
To make your local skill directory visible to the `gemini` command:
```bash
gemini skills link ./.agents/skills/go-expert
```

### Running with the Skill
When you're working in the CLI, you trigger the skill using a "slash command" or by explicitly invoking the persona:
```bash
# Ask a question using your specific skill context
gemini "/go-infra-expert how do I implement a NATS request-reply in this project?"
```

---

## 4. Pro Tip: The `GEMINI.md` Project File
If you want the CLI to *always* know about your skills without typing the slash command, create a `GEMINI.md` file in your root directory. This acts as a **persistent system prompt** for that specific folder.

```markdown
# Project Context
This project is a Go-based microservice using NATS. 
Always refer to `.agents/skills/go-expert/SKILL.md` for coding standards.
```

---

### Comparison for your Workflow

| Feature | Claude `skills.md` | Gemini CLI `SKILL.md` |
| :--- | :--- | :--- |
| **Discovery** | Manual upload/Project context | Auto-discovered in `~/.agents/skills` |
| **Metadata** | Often none (prose-based) | Requires YAML Frontmatter (`---`) |
| **Logic** | Textual constraints | Can be linked to local files/tools |

**Would you like me to take one of your existing Go or NATS skills and format it with the correct YAML frontmatter for the Gemini CLI?**
