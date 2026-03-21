# Final GitHub UI checklist (about 60 seconds)

Use this after repo files and API-visible settings are already in place.

## Required confirmation

1. Open **https://github.com/codethor0/dockercomms/settings/security_analysis**
2. Under **Code security and analysis**, confirm **Private vulnerability reporting** is **Enabled**. If it is off, turn it on and save.

## Optional (recommended if shown)

3. **Settings → General** (or **Releases** section, depending on UI): enable **Release immutability** if GitHub offers it for this account/repo.

## Optional (preference)

4. **Settings → General → Features**: **Projects** — disable if you want a leaner repo surface.
5. **Settings → General → Pull Requests**: **Allow auto-merge** — enable only if you want it alongside your branch protection rules.

## Done statement

When step 2 is confirmed (and optional steps decided):

> GitHub setup is complete for DockerComms, subject only to UI confirmation of private vulnerability reporting and optional release immutability.

See also: [GITHUB-SECURITY-SETUP.md](GITHUB-SECURITY-SETUP.md).
