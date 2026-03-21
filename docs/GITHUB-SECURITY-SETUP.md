# GitHub setup (Security and repo hygiene)

Some settings were applied via the GitHub REST API (repo admin token). Others require the web UI or a follow-up after the first Actions run.

## Applied via API (verify in Settings)

- Dependabot security updates: enabled (`dependabot_security_updates`)
- Dependabot vulnerability alerts: `PUT /repos/{owner}/{repo}/vulnerability-alerts` (204 when enabled)
- Secret scanning: already enabled; push protection enabled
- Private vulnerability reporting: sent in `security_and_analysis` PATCH (confirm under Settings → Code security and analysis; field may not appear in GET)
- Wiki: disabled
- Discussions: enabled
- Merge commits: disabled; squash and rebase: enabled
- Auto-delete head branches: enabled

## Applied in git (merge to `main`)

- `SECURITY.md`
- `.github/workflows/codeql.yml`
- `.github/dependabot.yml`
- `.github/ISSUE_TEMPLATE/*`
- `.github/pull_request_template.md`

## Branch protection on `main` (via API)

- Require status check: `test` (GitHub Actions app id 15368), strict
- Require pull request before merge with `required_approving_review_count: 0` (solo maintainer; still uses PR flow)
- Force pushes and branch deletion: disallowed

Repository administrators can bypass when **Do not enforce on administrators** is set in the classic rule (default when admins are not enforced on the rule UI equivalent). Confirm in Settings → Branches.

## After first CodeQL run on `main`

1. Open Settings → Branches → `main` protection.
2. Add required status check **Analyze** (workflow `CodeQL`, job `Analyze`) if not auto-listed, or use the exact name shown on a green run.
3. Under Security → Code scanning, confirm CodeQL is active.

## Manual (UI only or plan-dependent)

- **Dependabot alerts** (vulnerability alerts UI): For many public repos this appears once the dependency graph is active. If the Security tab still shows alerts off, enable **Dependabot alerts** under Code security and analysis.
- **Release immutability**: Repository or organization **Settings** → Releases / Packages; enable if your plan exposes it.
- **Projects**: Left enabled by default; ignore or archive if unused.

## Completion checklist

Use this after merging hardening to `main` and workflows have run:

- [ ] `SECURITY.md` on default branch
- [ ] Private vulnerability reporting shows enabled (Security tab)
- [ ] Dependabot (alerts and/or security updates) active
- [ ] Code scanning shows CodeQL runs
- [ ] Branch protection on `main` includes `test` and, after first run, `Analyze`
- [ ] Issue templates render on New Issue
- [ ] Discussions available if linked from issue config
