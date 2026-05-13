# Session Handoff Ledger

Updated: 2026-05-13T20:49:35.950Z
Current session: session-2026-05-13T20-49-35-912Z-5af4f328

## Outstanding Snapshots (3)

1. [pending] session-2026-04-14T00-08-49-595Z-4e78988f — feat/add-download-make-target (dirty)
   File: doc/.plan/session-handoff/sessions/session-2026-04-14T00-08-49-595Z-4e78988f.md
   Updated: 2026-04-14T00:08:49.595Z

2. [pending] session-2026-05-13T20-49-34-584Z-fe6e7d3b — main (clean)
   File: doc/.plan/session-handoff/sessions/session-2026-05-13T20-49-34-584Z-fe6e7d3b.md
   Updated: 2026-05-13T20:49:34.584Z

3. [pending] session-2026-05-13T20-49-35-912Z-5af4f328 — main (dirty)
   File: doc/.plan/session-handoff/sessions/session-2026-05-13T20-49-35-912Z-5af4f328.md
   Updated: 2026-05-13T20:49:35.912Z

## Recent Activity

- None

## Commands

Run these from the client project root (adjust $HOME/.config if you use a custom config home):

- `node "$HOME/.config/opencode/bin/session-handoff.mjs" list` — show pending snapshots
- `node "$HOME/.config/opencode/bin/session-handoff.mjs" ack <id> [--note "done"]` — mark complete
- `node "$HOME/.config/opencode/bin/session-handoff.mjs" dismiss <id> --reason "why"` — abandon work
- `node "$HOME/.config/opencode/bin/session-handoff.mjs" write --trigger "/pro:session.handoff"` — capture a fresh snapshot

If you vendor the CLI into a repo instead:

- `node bin/session-handoff.mjs list|ack|dismiss|write ...`

Compatibility note: some older snapshot files may still mention `node bin/session-handoff.mjs ...`. If your repo does not contain that file, use the globally installed CLI commands listed above.

All snapshots live under `doc/.plan/session-handoff/sessions/`. Review each file before acknowledging or dismissing it.
