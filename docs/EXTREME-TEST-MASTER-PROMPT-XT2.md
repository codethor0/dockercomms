# Master prompt: XT-2 — Live Docker + full-matrix execution

**Version: XT-2** — next step after [XT-1](EXTREME-TEST-MASTER-PROMPT.md). **Does not** supersede [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11) for GA/closure; **does** add Docker-first live runs and **anti-summary** rules. Use [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md) for principles.

---

## Hard truths (keep the prompt honest)

- **“Absolutely no bugs”** is not something any agent or test suite can prove; the strongest defensible claim is **no known failures in the matrix you actually ran**, with logs.
- **Cursor** can only use Docker if the environment exposes it (Docker Desktop, Linux daemon, or Dev Containers with the socket). The prompt must **fail closed** when Docker is missing.
- **“Do everything in Docker”** can still need a **host** step (e.g. `docker-e2e` gates may produce a **Linux** binary; **macOS** may need `make clean && make build` for a **Mach-O** host run — see the runbook and this document’s **GATE 0** / **C** / **D** flow).

---

## Relationship to other prompts

| Prompt / doc | Role |
|--------------|------|
| **v11** / [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) | GA, §A/§B, NOT RUN — **releases** |
| [EXTREME-TEST-MASTER-PROMPT.md](EXTREME-TEST-MASTER-PROMPT.md) (XT-1) | Full adversarial plan |
| **XT-2** (this file) | **Execution** + **Docker-first** + **anti-summary** (first output = shell + real command + real output) |
| [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md) | Oracle, pyramid, repro bar, bash for NEG |

**“Open Docker in Cursor”** = your machine’s Docker Desktop (or a Dev Container with the Docker socket). The agent runs `docker` in the **terminal**; there is no separate “Cursor-only Docker” unless you configure Dev Containers or a remote. XT-2 expects the same `docker` CLI as a normal shell.

---

## Paste block (for Cursor) — everything in the fence

```
MASTER PROMPT: dockercomms — XT-2 “Live Docker + execution evidence” (not a GA/closure runbook)
Use with: local clone of `codethor0/dockercomms`. Read `docs/TESTING-DOCTRINE.md`, `docs/EXTREME-TEST-MASTER-PROMPT.md` (XT-1), and `docs/RELEASE-RUNBOOK.md` (v11) first. XT-2 **adds** Docker-first live runs and **anti-summary** rules. It does **not** replace **GA/§A/§B** evidence — use the runbook for that.
MISSION
Run **real** commands, **real** Docker, **real** local registry, and (if creds exist) **real** GHCR paths. **Maximize** evidence; **minimize** narrative. Goal: **zero unexpected product failures** in the **defined matrix** for this SHA, not a proof of “no bugs in the universe.”
ANTI-SUMMARY RULE (HARD)
- The first 20 lines of your output **must** be: **shell header** (UTC, path, SHA, Docker status) + **the first command you ran** + **its output** (or exit + one-line error).  
- **Forbidden** as an opening: long plans, “here’s what I would do,” or executive summary **before** any command output.  
- After each phase, **at most 3** sentences of interpretation; then **next** commands.  
- If you cannot run a step, **one line** `BLOCKED: {reason}` and move on — do not pad.
DOCKER AVAILABILITY (GATE 0)
1. `docker info` (or `docker version`) — if **fails**, state **BLOCKED: Docker not available in this environment** and run **only** host `go test` + static checks; **do not** claim “full XT-2” complete.  
2. If **OK**, one line: **Docker: OK** + **context** (e.g. `docker context` name).  
3. For **E2E in container** (project script): from repo root run `./scripts/docker-e2e.sh gates` (or as documented). If script **fails**, capture **last 100 lines** of output (§8b style).  
4. **After** any step that overwrites **root** `./dockercomms` with a **Linux** build: `make clean && make build` on **host** + `file` one line — **mandatory** before further **host** CLI tests.
LIVE TEST MATRIX (RUN IN ORDER; **STOP** only on **unexpected** PRODUCT failure unless user said “continue and log all”)
**A. Host static + unit (always)**  
`gofmt -l .` (empty or list files) · `go vet ./...` · `go test ./...` · `go test -race ./...` · `golangci-lint run ./...` (if project standard) · `make coverage-gate`  
**B. Script preflight (no network OR check-only as designed)**  
`./scripts/run-integration.sh --check` · `./scripts/login-and-run-integration.sh --check`  
**C. Container gates (if Docker OK)**  
`./scripts/docker-e2e.sh` **gates** → then **host rebuild** per above  
**D. Local registry E2E (isolated, bash driver)**  
`docs/repro.md` (or `RELEASE-RUNBOOK`): **one** **happy** round-trip per **isolation_id**; **two** **NEG** (expected non-zero) with `Expected` from **README** / doc **section heading**; **two** **flake** iterations on **different** `isolation` keys  
**E. Optional matrix (if user or time says)**  
- `E2E_LINUX_MATRIX=1` in user message: `./scripts/linux-distro-matrix.sh` if present, else `NOT RUN`  
- `GH_PAT` or `~/.dockercomms_gh_pat` on **host**: `docs/FINAL-RELEASE-GO-NO-GO.md` **Section B** by heading, then `./scripts/login-and-run-integration.sh` (full), then `./scripts/docker-e2e.sh` **integration** → **cli** → **full** — if **no** creds: run **integration** once, **quote** skip, label **NOT RUN (auth)** — not “pass”
CLEAN CODE + NO REGRESSIONS (if code changes authorized in **this** message)
- **Smallest** diff; **one** issue per change or tight cluster.  
- **gofmt**; **tests** in same package for behavior change.  
- Run at minimum: `go test -race` on **touched** packages and `go test ./...` if fast enough.  
- **No** disabling linters, **no** weakening tests, **no** `// nolint` without justification in the same PR.  
- **No** “cleanup” in the same commit as a bugfix **unless** trivial.
ORACLES
- `README` **## Exit Codes**; `docs/repro.md`; `SPEC`/`ARCH` as needed. Mismatch = **bug** (code) or **DOC-DEBT** (docs) — state which.
REPORT (END — ONLY AFTER COMMANDS)
1. **SHA** (3-way if claiming published tip), **clean/dirty** tree.  
2. **Table**: Phase | Command | Result | Note (isolation_id / NOT RUN / B-UX-HOST).  
3. **Unexpected PRODUCT failures** (must be **empty** for a “clean” claim): or **list** with **5-field** repro (command, isolation_id, expected, actual, first error line).  
4. **BLOCKED** rows for Docker / auth / time.  
5. **One paragraph**: what is **not** covered (GHCR, chaos, huge files) — not faux humility, **concrete** gaps.  
6. If user asked **“no bugs”** language: use **one** sentence: **“No *unexpected* failures in the above matrix; not a guarantee of total correctness.”**
USER MESSAGE SHOULD INCLUDE
- `Clone: /path/to/dockercomms`  
- `CODE_CHANGES_ALLOWED=yes|no` and **if** `yes` **one** line scope.  
- `Time budget: e.g. 2h` or `run_all_phases=yes`  
- `GH_PAT present on host: yes|no` (not the value)  
- Optional: `E2E_LINUX_MATRIX=1`, `CONTINUE_ON_FAILURE=yes` to log all failures in one go.
BANNED
- Opening with summary; claiming green without output; “absolutely no bugs” globally; skipping Docker gate when `docker info` works; P1 on shared-inbox without isolation retry; secrets in the log; emojis (optional).
END XT-2
```

---

## How to use in Cursor (paste under the block)

One line, for example:

`Clone: /path/to/dockercomms` · `CODE_CHANGES_ALLOWED=no` · `Time budget: 2h` · `GH_PAT present on host: no`

Change `CODE_CHANGES_ALLOWED=yes` only when you want the agent to land fixes; keep one focused PR or commit per bug cluster.

---

## See also

- [TESTING-DOCTRINE.md](TESTING-DOCTRINE.md)
- [EXTREME-TEST-MASTER-PROMPT.md](EXTREME-TEST-MASTER-PROMPT.md) (XT-1)
- [RELEASE-RUNBOOK.md](RELEASE-RUNBOOK.md) (v11 / operator GA)
