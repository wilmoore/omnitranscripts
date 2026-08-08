# 0005. Separate CLI Data and Diagnostic Streams

Date: 2026-08-07

## Status

Accepted

## Context

The transcription CLI mixed progress messages, headings, segment metadata, and transcript text on standard output. Consumers therefore could not safely pipe the transcript to clipboard tools, files, or text-processing commands without also capturing diagnostic content.

## Decision

The transcribe command treats standard output as its data interface and writes only the transcript there. Usage text, progress, errors, and completion statistics are written to standard error. The Makefile wrapper follows the same contract so it does not contaminate the CLI output stream.

## Consequences

Shell pipelines can consume transcript text directly while diagnostics remain visible by default. Consumers that need a fully quiet command can redirect standard error. Human-readable segment and summary output is no longer included in standard output.

## Alternatives Considered

- Add a dedicated clipboard target, which would duplicate platform-specific shell behavior and reduce composability.
- Keep mixed output and require downstream filtering, which would make the output format fragile and difficult to automate.
- Hide diagnostics entirely, which would make long-running transcription work harder to observe and troubleshoot.

## Related

- Planning: `.plan/.done/chore-fix-clipboard-transcribe-output/`
