# Testing doctrine (project principles)

Short, reusable. Operator closure / GA evidence uses **`docs/RELEASE-RUNBOOK.md`** and the **v11**-style master prompt, not this page.

**Oracle first** — Every behavior you test has an **expected** from docs (`README` exit codes, `SPEC.md`, `docs/repro.md`) or from the code; if neither is clear, file **docs debt** before calling it a "bug."

**Pyramid** — Many fast **unit** / **property** tests; some **integration**; few heavy **E2E**; rare **extreme** / chaos runs (expensive, dedicated sessions only).

**Host vs Docker** — If **`docker info` fails**, treat Docker-dependent work (script `--check` that requires a daemon, `docker-e2e` **gates**, local `registry:2`, container GHCR paths) as **BLOCKED**; **host** static + unit Go (`gofmt`, `go test`, `go vet`, lint, coverage) can still be **valid signal** for that checkout (see **XT-2** [GATE 0](EXTREME-TEST-MASTER-PROMPT-XT2.md)).

**Risk order** — Crypto, path safety, auth to registry, "verify before materialize," and **exit-code contracts** before cosmetic UX.

**Repro before label** — No **P1** without **command** + **environment** + **isolation key** (I7 / unique OCI repo) and **first error line**.

**Surgical fixes** — One bug (or one tight cluster) per change; tests or regression lock before/after; no refactors in the same PR as a hotfix unless unavoidable.

**Bash for exit capture** on expected failures; see `docs/RELEASE-RUNBOOK.md` (`zsh` can confuse `errexit` with NEG cases).

**LLM is the runner/analyst, not the judge of green** — Green = you **saw** the command output; the model never invents pass/fail.

**Stability** — **Flake** = two **isolated** successes (I8); if only flaky under load, that is a separate **concurrency/timeout** class.

**Fuzz is targeted** — Use Go fuzz on parsers, paths, digests, filenames; not random UI clicks.

**Stop rule** — End a session with a **written** list: **fixed** | **repro'd open** | **WONT_FIX / docs** | **deferred (with reason).**

## Related

- Adversarial / extreme QA: **`docs/EXTREME-TEST-MASTER-PROMPT.md`** (XT-1); live Docker + anti-summary execution: **`docs/EXTREME-TEST-MASTER-PROMPT-XT2.md`** (XT-2); change-aware / PR / merge confidence: **`docs/EXTREME-TEST-MASTER-PROMPT-XT3.md`** (XT-3).
- Release closure, NOT RUN, GA: **`docs/RELEASE-RUNBOOK.md`**.
