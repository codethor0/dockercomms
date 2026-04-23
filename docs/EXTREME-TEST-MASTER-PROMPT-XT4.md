# Master prompt: XT-4 — Evidence-based test and triage (agent handoff)

**Version: XT-4** — maximal, **evidence-first** pass: real shell output, no fabricated pass/fail. **Does not** replace [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11) for GA/§A/§B.

**Next step:** [EXTREME-TEST-MASTER-PROMPT-XT5.md](EXTREME-TEST-MASTER-PROMPT-XT5.md) (XT-5) adds **preflight Docker** + **STATIC_ONLY** / **DOCKER_REQUIRED** so you do not spend a long Phase 1 before discovering Docker is down when you needed E2E.

---

## Non-negotiable truth rules (summary)

- Do not claim “no bugs” / 100% correct; strongest honest claim: no **unexpected** failures in the **matrix run** for this checkout, OS, Docker, and time budget.
- Do not fabricate results; **NOT RUN** with one-line reason if not executed.
- Do not print secrets; **GH_PAT** = PRESENT or ABSENT only.
- **NEG** / exit capture: **bash** driver (see [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md)); if `docker info` fails, Docker-dependent phases are **BLOCKED**; **host** Phase 1 Go static + unit can still be valid ([TESTING-DOCTRINE.md](TESTING-DOCTRINE.md), [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) GATE 0).
- After `docker-e2e.sh` **gates** (or any step that can overwrite the root binary with a Linux build): **`make clean && make build`** + **`file ./dockercomms`** on host before further host CLI tests.

## Phases (typical order)

| Phase | Content |
|-------|---------|
| **0** | XT-3 pre-flight: what changed; targeted `go test -race` on touched `pkg/` / `cmd/` / `internal/` |
| **1** | Static + unit: `gofmt -l`, `go vet`, `go test`, `go test -race`, `golangci-lint`, `make coverage-gate` |
| **2** | `./scripts/run-integration.sh --check`, `./scripts/login-and-run-integration.sh --check` |
| **3** | `./scripts/docker-e2e.sh gates` → host rebuild + `file` |
| **4** | Local registry E2E (bash): happy path, NEG1/NEG2, I8 two isolated flakes per [repro.md](repro.md) |
| **5** | Optional GHCR: PAT present → Section B + `login-and-run` + `docker-e2e` integration → cli → full; else skip quote + **NOT RUN (auth)** |
| **6** | Optional: `linux-distro-matrix.sh`, chaos — time / `EXTREME` |

## Output template (end of session)

Use sections such as:

- **### XT-4 RESULT — dockercomms** — SHA, 3-way / clean / Docker / time budget
- **### Command outcomes (table)** — Phase | Command | Exit | Note
- **### Unexpected product failures** — empty for “clean” or 5-field rows per [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md)
- **### NOT RUN / BLOCKED**
- **### P0–P1 register** / **P2–P3 / DOC-DEBT**
- **### HONESTY LINE** — one sentence: no unexpected failures in matrix above; not total correctness

**Opening (anti-summary):** first block = `date -u`, `pwd`, `git status`, SHAs, `go version`, `docker` check, **then** first command output — no plan wall before evidence.

---

## Related

- [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md) · [XT-1](EXTREME-TEST-MASTER-PROMPT.md) · [XT-2](EXTREME-TEST-MASTER-PROMPT-XT2.md) · [XT-5](EXTREME-TEST-MASTER-PROMPT-XT5.md)
- [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11 / GA)
