# multitab

**Your AI agents are wasting 80% of your CI builds. Fix it in 30 seconds.**

If you run multiple Claude Code sessions (or any AI coding agents) in the same repo, you already know the pain: five agents pushing independently to `main`, five Vercel builds triggered, four of them cancelled. Expensive. Chaotic. Unnecessary.

multitab turns your terminal into a mission control dashboard. One view. One coordinated push. Zero wasted builds.

```
╔══ MULTITAB ═══════════════════════════════════════════════════╗
║                                                               ║
║  AGENTS                     STATUS       COMMITS   FILES      ║
║  ──────────────────────────────────────────────────────────── ║
║  ● fix-history              ✅ STAGED     2         4         ║
║  ● expire-requests          ✅ STAGED     1         2         ║
║  ● quick-booking            🔨 WORKING    3         7         ║
║  ○ fix-schedule-api         💤 IDLE       0         0         ║
║                                                               ║
║  DEPLOY QUEUE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 2/4      ║
║  ██████████████████████░░░░░░░░░░░  50% ready                 ║
║                                                               ║
║  [P]ush  [D]iff  [R]efresh  [L]og  [Q]uit                   ║
╚═══════════════════════════════════════════════════════════════╝
```

## Why

We build with multiple AI agents running in parallel — each in its own git worktree, each doing real work. The problem was never the agents. The problem was the push.

Every independent push triggers a full CI pipeline. Most get cancelled when the next one lands seconds later. That's money, time, and build minutes evaporating into nothing.

So we built multitab. Not because the world needed another dev tool — but because *we* needed to stop fighting our own infrastructure.

## What it does

- **Auto-detects worktrees** — scans `.claude/worktrees/` and any git worktrees in the repo
- **Shows agent status** — WORKING (active changes), STAGED (merged to local main), IDLE (nothing happening)
- **Lists staged commits** — everything on local main ahead of `origin/main`, ready to ship
- **Detects conflicts** — flags when two agents touched the same file before you push
- **Detects migrations** — warns if staged work includes database migrations
- **Push with live progress** — animated step-by-step: fetch → rebase → build → push, with per-step timing
- **Auto-refreshes** — dashboard updates every 5 seconds, no manual polling

## Install

```bash
# From source (requires Go 1.21+)
go install github.com/goldenfocus/multitab@latest

# Or clone and build
git clone https://github.com/goldenfocus/multitab.git
cd multitab
go build -o multitab .
```

## Usage

```bash
multitab              # launch the TUI dashboard
multitab status       # quick text-only status (pipe-friendly)
multitab push         # non-interactive batch push (for scripts/CI)
multitab init         # create .multitab.toml config (first run)
```

### The dashboard

Run `multitab` inside any git repo. It discovers all worktrees, inspects their state, and renders the dashboard. Press `P` to push everything staged on local main in one coordinated deploy.

### Text-only status

```bash
$ multitab status

MULTITAB STATUS
═══════════════════════════════════════════════════
AGENT                        STATUS         COMMITS    FILES
────────────────────────────────────────────────────
✓ fix-history                STAGED         2          4
• expire-requests            WORKING        1          2
○ fix-schedule-api           IDLE           0          0

STAGED COMMITS (2):
  a1b2c3d fix(history): correct pagination offset
  e4f5g6h fix(history): handle empty state

DEPLOY QUEUE: 1/3 ready
CONFLICTS: None
```

### Configuration

`multitab init` creates a `.multitab.toml` at the repo root:

```toml
[repo]
main_branch = "main"
worktree_dir = ".claude/worktrees"

[build]
command = "bash scripts/safe-build.sh"    # auto-detected
timeout = "10m"

[push]
pre_push_hooks = true
auto_cleanup = true

[tui]
refresh_interval = "5s"
```

## Built with

- [Bubbletea](https://github.com/charmbracelet/bubbletea) — the Elm Architecture for terminals
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — styling that makes terminals feel premium
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- Pure Go git operations — no `go-git` dependency, just `os/exec` calling the git you already have

## The philosophy

The best tools disappear into your workflow. You shouldn't have to *think* about coordinating pushes — you should just see what's ready, press P, and watch it fly.

We believe AI agents are going to be a normal part of how software gets built. Not someday — right now. And the tooling around them should be as thoughtful and beautiful as the code they help us write.

multitab is a small piece of that future. A calm, clear window into the organized chaos of multi-agent development.

## Roadmap

- [ ] CI integration (read Vercel/GitHub Actions build status)
- [ ] Post-push cleanup (auto-remove merged worktrees)
- [ ] Custom themes
- [ ] Multi-repo support
- [ ] Homebrew formula (`brew install multitab`)

## License

[MIT](LICENSE) — use it, fork it, make it yours.

---

*Built with love and too many terminal tabs by [GoldenFocus](https://github.com/goldenfocus).*
