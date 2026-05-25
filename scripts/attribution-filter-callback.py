#!/usr/bin/env python3
"""git-filter-repo commit callback: humanize dependabot authors and strip agent trailers."""

def commit_callback(commit, metadata):
    if commit.author_name == b"dependabot[bot]":
        commit.author_name = b"Isaac Thor"
        commit.author_email = b"codethor@gmail.com"
    if commit.committer_name == b"dependabot[bot]":
        commit.committer_name = b"Isaac Thor"
        commit.committer_email = b"codethor@gmail.com"

    if not commit.message:
        return
    text = commit.message.decode("utf-8", errors="replace")
    kept = []
    for line in text.splitlines():
        if line.startswith("Co-authored-by: Cursor"):
            continue
        if line.startswith("Co-authored-by: dependabot"):
            continue
        if line.startswith("Made-with: Cursor"):
            continue
        kept.append(line)
    body = "\n".join(kept).rstrip()
    commit.message = (body + "\n").encode("utf-8") if body else b"chore: dependency update\n"
