# Pinned issue: "Start here: What DockerComms does and how to run tests"

Create a new issue with this title and paste the body below.

---

## Start here: What DockerComms does and how to run tests

Thanks for checking out DockerComms.

At a high level:

- DockerComms is an OCI-native secure file transport CLI.
- It sends and receives files as OCI artifacts via standard registries (GHCR, Docker Hub, GCR, etc.).
- It enforces verify-before-materialize and strict path sanitization so untrusted payloads never get written to disk before verification.

If you are new to the project, the three things you probably want first are:

1. **What the tool does**  
   See the "What is DockerComms?" section in `README.md` for the security model, protocol overview, and key invariants (verify-before-materialize, tag grammar, exit codes).

2. **How to run tests locally (no credentials required)**  
   From the repo root:

   ```bash
   go test ./...
   go test -race ./...
   golangci-lint run ./...
   make coverage-gate
   ./scripts/run-integration.sh --check
   ./scripts/login-and-run-integration.sh --check
   ```

   These gates should all be green before you open a PR.

3. **How to run GHCR integration safely (optional)**  
   If you want to exercise live registry behavior:

   - Set `GH_USER`, `GH_PAT`, `DOCKERCOMMS_IT_GHCR_REPO`, and `DOCKERCOMMS_IT_RECIPIENT` in your shell.
   - Run: `./scripts/login-and-run-integration.sh`

   The scripts:
   - Use `docker login ghcr.io` non-interactively with `--password-stdin`
   - Never echo the PAT
   - Use `umask 077` when touching secret files
   - Offer `--check` modes that do validation without network calls

   There is also a Docker E2E harness (`scripts/docker-e2e.sh`) that can run the same gates and integration flows inside a container. See `README.md` for details.

**If you are planning to contribute:**

- Run the local gates before opening a PR.
- If your changes affect integration behavior or CLI flows, mention which scripts/tests you ran (e.g., `login-and-run-integration.sh`, `docker-e2e.sh cli`) in your PR description.

Questions, bug reports, or ideas are welcome. Please include the exact command you ran and the full error output when filing an issue.
