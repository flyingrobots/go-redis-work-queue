#!/usr/bin/env python3
"""Extract comments by one author from a GitHub pull request using ``gh``.

Examples:
  python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit \
      --out docs/audits/coderabbit-pr123-comments.md

  python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit \
      --prompts-only \
      --prompts-out docs/audits/coderabbit-pr123-prompts.md

If ``--pr`` is omitted, the pull request associated with the current branch is
used. Output defaults to ``docs/audits/``. The GitHub CLI must be installed and
authenticated for the current repository.
"""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
import json
from pathlib import Path
import re
import subprocess
from typing import Any


Comment = dict[str, Any]


def run(command: list[str]) -> str:
    """Run a command and return its trimmed standard output."""
    return subprocess.check_output(command, text=True).strip()


def gh_json(arguments: list[str]) -> Any:
    """Run a ``gh`` command and decode its JSON response."""
    return json.loads(run(arguments))


def gh_paginated_list(endpoint: str) -> list[Comment]:
    """Fetch and flatten every list page returned by a GitHub REST endpoint."""
    pages = gh_json(["gh", "api", endpoint, "--paginate", "--slurp"])
    if not isinstance(pages, list):
        raise SystemExit(f"Expected paginated JSON list from {endpoint}")

    items: list[Comment] = []
    for page in pages:
        if not isinstance(page, list):
            raise SystemExit(f"Expected each page from {endpoint} to be a list")
        for item in page:
            if not isinstance(item, dict):
                raise SystemExit(f"Expected comment objects from {endpoint}")
            items.append(item)
    return items


def parse_timestamp(value: str) -> datetime:
    """Return a timezone-aware timestamp suitable for stable sorting."""
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00"))
    except (AttributeError, ValueError):
        return datetime.min.replace(tzinfo=timezone.utc)


def matching_comments(
    issue_comments: list[Comment],
    review_comments: list[Comment],
    reviews: list[Comment],
    author_substring: str,
) -> list[Comment]:
    """Normalize and chronologically sort comments matching an author."""
    matches: list[Comment] = []

    def add(kind: str, comment: Comment) -> None:
        user = str((comment.get("user") or {}).get("login", ""))
        if author_substring not in user.lower():
            return

        body = str(comment.get("body") or "")
        if kind == "review" and not body:
            return

        matches.append(
            {
                "kind": kind,
                "created": comment.get("created_at")
                or comment.get("submitted_at")
                or "",
                "user": user,
                "body": body,
                "path": comment.get("path") or "",
                "line": comment.get("line")
                or comment.get("original_line")
                or "",
                "url": comment.get("html_url")
                or comment.get("pull_request_url")
                or "",
            }
        )

    for comment in issue_comments:
        add("issue_comment", comment)
    for comment in review_comments:
        add("review_comment", comment)
    for review in reviews:
        add("review", review)

    matches.sort(key=lambda item: parse_timestamp(str(item["created"])))
    return matches


def prompts_from(body: str) -> list[str]:
    """Extract fenced or inline ``Prompt for AI Agents`` content."""
    prompts: list[str] = []
    fenced_spans: list[tuple[int, int]] = []
    fenced_pattern = re.compile(
        r"prompt\s*for\s*ai\s*agents[^`]*```(?:[^\n]*)\n?(.*?)```",
        re.IGNORECASE | re.DOTALL,
    )
    for match in fenced_pattern.finditer(body):
        prompt = match.group(1).strip()
        if prompt and prompt not in prompts:
            prompts.append(prompt)
        fenced_spans.append(match.span())

    inline_pattern = re.compile(
        r"prompt\s*for\s*ai\s*agents\s*:?\s*(.+?)"
        r"(?:\n\s*\n|\n#+|\n>\s|$)",
        re.IGNORECASE | re.DOTALL,
    )
    for match in inline_pattern.finditer(body):
        if any(start <= match.start() < end for start, end in fenced_spans):
            continue
        prompt = match.group(1).strip()
        if prompt and prompt not in prompts:
            prompts.append(prompt)

    return prompts


def write_heading(handle: Any, title: str, repo: str) -> None:
    handle.write(f"# {title}\n\n")
    handle.write(f"Repo: {repo}\n\n")


def write_item(handle: Any, item: Comment, body: str) -> None:
    location = f" ({item['path']}:{item['line']})" if item["path"] else ""
    handle.write(
        f"- [{item['kind']}] {item['created']} by {item['user']}"
        f"{location}\n\n"
    )
    handle.write(f"{body.rstrip()}\n\n")
    if item["url"]:
        handle.write(f"URL: {item['url']}\n\n")


def write_comments(
    path: Path,
    repo: str,
    pr_number: int,
    author: str,
    items: list[Comment],
) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        write_heading(
            handle,
            f"Review Comments Matching {author!r} for PR #{pr_number}",
            repo,
        )
        for item in items:
            write_item(handle, item, str(item["body"]))


def write_prompts(
    path: Path,
    repo: str,
    pr_number: int,
    author: str,
    items: list[Comment],
) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        write_heading(
            handle,
            f"Review Prompts Matching {author!r} for PR #{pr_number}",
            repo,
        )
        for item in items:
            for prompt in prompts_from(str(item["body"])):
                write_item(handle, item, prompt)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pr", type=int, help="pull request number")
    parser.add_argument(
        "--author",
        default="coderabbit",
        help="case-insensitive substring of the author login",
    )
    parser.add_argument("--out", help="output Markdown path")
    parser.add_argument(
        "--prompts-only",
        action="store_true",
        help='extract only "Prompt for AI Agents" sections',
    )
    parser.add_argument("--prompts-out", help="prompts-only Markdown output path")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    repo = run(["gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner"])

    if args.pr is not None:
        pr_number = args.pr
    else:
        try:
            pr_number = int(run(["gh", "pr", "view", "--json", "number", "-q", ".number"]))
        except (subprocess.CalledProcessError, ValueError) as error:
            raise SystemExit(
                "No PR number supplied and no PR is associated with the current branch."
            ) from error

    author_substring = args.author.lower()
    issue_comments = gh_paginated_list(f"repos/{repo}/issues/{pr_number}/comments")
    review_comments = gh_paginated_list(f"repos/{repo}/pulls/{pr_number}/comments")
    reviews = gh_paginated_list(f"repos/{repo}/pulls/{pr_number}/reviews")
    items = matching_comments(
        issue_comments,
        review_comments,
        reviews,
        author_substring,
    )

    if args.prompts_only:
        output = Path(
            args.prompts_out
            or f"docs/audits/{author_substring}-pr{pr_number}-prompts.md"
        )
        write_prompts(output, repo, pr_number, args.author, items)
    else:
        output = Path(
            args.out
            or f"docs/audits/{author_substring}-pr{pr_number}-comments.md"
        )
        write_comments(output, repo, pr_number, args.author, items)

    print(output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
