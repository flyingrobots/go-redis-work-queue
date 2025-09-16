#!/usr/bin/env python3
"""
Generate a Code Review Worksheet from CodeRabbit "Prompt for AI Agents" comments on a GitHub PR.

Requirements:
- GitHub CLI (`gh`) installed and authenticated to the target repo.

Usage examples:
  python3 scripts/generate_cr_worksheet.py --pr 3
  python3 scripts/generate_cr_worksheet.py --repo flyingrobots/go-redis-work-queue --pr 3 \
      --author coderabbit --agent CodeRabbit

Outputs a Markdown worksheet to: docs/audits/code-reviews/PR{pr}/{head_sha}.md

The worksheet format matches the in-repo example, including:
- Front matter and header table (Date, Agent, SHA, Branch link, PR link)
- Accepted/Rejected templates
- All prompt items with checkboxes, fenced prompt text, and a `{response}` placeholder
- A horizontal rule and a Conclusion section at the end

Notes:
- Filters for comments authored by `--author` (substring match; default: coderabbit)
- Extracts only fenced code blocks that follow a "Prompt for AI Agents" section header to avoid HTML-only artifacts
- Skips empty/HTML-only segments (e.g., stray `</summary>`)
"""
import argparse
import datetime as dt
import json
import os
import re
import subprocess
from pathlib import Path


def sh_json(args: list[str]):
    out = subprocess.check_output(args, text=True)
    return json.loads(out)


def sh_str(args: list[str]) -> str:
    return subprocess.check_output(args, text=True).strip()


def detect_repo() -> str:
    return sh_str(["gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner"])


def fetch_pr(repo: str, pr: int) -> dict:
    return sh_json(["gh", "api", f"repos/{repo}/pulls/{pr}"])


def fetch_comments(repo: str, pr: int) -> tuple[list[dict], list[dict], list[dict]]:
    issue_comments = sh_json(["gh", "api", f"repos/{repo}/issues/{pr}/comments", "--paginate"])  # convo
    review_comments = sh_json(["gh", "api", f"repos/{repo}/pulls/{pr}/comments", "--paginate"])  # inline
    reviews = sh_json(["gh", "api", f"repos/{repo}/pulls/{pr}/reviews", "--paginate"])  # bodies
    return issue_comments, review_comments, reviews


def is_from_author(item: dict, author_sub: str) -> bool:
    user = (item.get("user") or {}).get("login", "")
    return author_sub in user.lower()


def extract_prompts_from_body(body: str) -> list[str]:
    if not body:
        return []
    txt = body
    prompts: list[str] = []
    # Normalize to reduce false negatives
    hay = txt.lower()
    if "prompt for ai agents" not in hay:
        return []

    # Strategy 1: capture fenced blocks following the header text
    # Example: ... Prompt for AI Agents ... ```<lang>?\n<block>\n```
    for m in re.finditer(r"(?is)prompt\s*for\s*ai\s*agents.*?```[a-zA-Z0-9_\-]*\s*(.*?)```", txt):
        block = m.group(1).strip()
        if block:
            prompts.append(block)

    # Strategy 2: if none found, capture inline text up to a blank line/HTML close
    if not prompts:
        m = re.search(r"(?is)prompt\s*for\s*ai\s*agents\s*:?\s*(.+?)\n\s*\n", txt)
        if m:
            cand = m.group(1).strip()
            # Remove trivial HTML-only artifacts
            cand_nohtml = re.sub(r"(?is)<[^>]+>", "", cand).strip()
            if len(cand_nohtml) >= 10:
                prompts.append(cand)
    # Filter out HTML-only or trivial closers like </summary>
    cleaned: list[str] = []
    for p in prompts:
        p2 = re.sub(r"(?is)<[^>]+>", "", p).strip()
        if not p2:
            continue
        if re.fullmatch(r"(?is)</?summary>\s*", p2):
            continue
        cleaned.append(p)
    return cleaned


def build_title(item: dict) -> str:
    path = item.get("path") or ""
    line = item.get("line") or item.get("original_line") or ""
    if path and line:
        return f"{path}:{line}"
    if path:
        return path
    # Fallback
    return f"{item.get('kind','review')} {item.get('created','')}"


def render_header(date_str: str, agent: str, sha: str, branch: str, branch_url: str, pr_num: int, pr_url: str, title_filename: str, repo_full: str) -> str:
    return """---
title: {title_filename}
description: Preserved review artifacts and rationale.
audience: [contributors]
domain: [quality]
tags: [review]
status: archive
---

# Code Review Feedback

| Date | Agent | SHA | Branch | PR |
|------|-------|-----|--------|----|
| {date} | {agent} | `{sha}` | [{branch}]({branch_url} "{repo_branch}") | [PR#{pr}]({pr_url}) |

## Instructions

Please carefully consider each of the following feedback items, collected from a GitHub code review.

Please act on each item by fixing the issue, or rejecting the feedback. Please update this document and fill out the information below each feedback item by replacing the text surrounded by curly braces. 

### Accepted Feedback Template

```markdown

> [!note]- **Accepted**
> | Confidence | Remarks |
> |------------|---------|
> | {{confidence_score_out_of_10}} | {{confidence_rationale}} |
>
> ## Lesson Learned
> 
> {{lesson}}
>
> ## What did you do to address this feedback?
>
> {{what_you_did}}
>
> ## Regression Avoidance Strategy
>
> {{regression_avoidance_strategy}}
>
> ## Notes
>
> {{any_additional_context_or_say_none}}

```

### Rejected Feedback Template

Please use the following template to record your rejections.

```markdown

> [!CAUTION]- **Rejected**
> | Confidence | Remarks |
> |------------|---------|
> | {{confidence_score_out_of_10}} | {{confidence_rationale}} |
>
> ## Rejection Rationale
>
> {{rationale}}
>
> ## What you did instead
>
> {{what_you_did}}
>
> ## Tradeoffs considered
>
> {{pros_and_cons}}
>
> ## What would make you change your mind
>
> {{change_mind_conditions}}
>
> ## Future Plans
>
> {{future_plans}}

```

---

## CODE REVIEW FEEDBACK

The following section contains the feedback items, extracted from the code review linked above. Please read each item and respond with your decision by injecting one of the two above templates beneath the feedback item.
""".format(
        title_filename=title_filename,
        date=date_str,
        agent=agent,
        sha=sha,
        branch=branch,
        branch_url=branch_url,
        repo_branch=f"{repo_full}:{branch}",
        pr=pr_num,
        pr_url=pr_url,
    )


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", help="owner/name; defaults to current repo via gh")
    ap.add_argument("--pr", type=int, required=True, help="Pull request number")
    ap.add_argument("--author", default="coderabbit", help="Substring match for author login (default: coderabbit)")
    ap.add_argument("--agent", default="CodeRabbit", help="Agent label for header (default: CodeRabbit)")
    ap.add_argument("--out-root", default="docs/audits/code-reviews", help="Root dir to write worksheets under PR{n}/")
    args = ap.parse_args()

    repo = args.repo or detect_repo()
    pr = fetch_pr(repo, args.pr)
    head_sha = pr.get("head", {}).get("sha", "")
    branch = pr.get("head", {}).get("ref", "")
    repo_full = pr.get("head", {}).get("repo", {}).get("full_name", repo)
    pr_url = pr.get("html_url", f"https://github.com/{repo}/pull/{args.pr}")
    branch_url = f"https://github.com/{repo_full}/tree/{branch}"

    issue_comments, review_comments, reviews = fetch_comments(repo, args.pr)
    author_sub = args.author.lower()

    # Collect items with prompts
    collected = []  # list of dict(title, prompt)
    seen = set()
    def add_item(item: dict, body: str):
        prompts = extract_prompts_from_body(body)
        for p in prompts:
            title = build_title(item)
            key = (title, p.strip())
            if key in seen:
                continue
            seen.add(key)
            collected.append({"title": title, "prompt": p.strip()})

    for c in issue_comments:
        if is_from_author(c, author_sub):
            c["kind"] = "issue_comment"
            add_item(c, c.get("body") or "")
    for c in review_comments:
        if is_from_author(c, author_sub):
            c["kind"] = "review_comment"
            add_item(c, c.get("body") or "")
    for r in reviews:
        if is_from_author(r, author_sub) and r.get("body"):
            r["kind"] = "review"
            add_item(r, r.get("body") or "")

    # Prepare output path
    out_dir = Path(args.out_root) / f"PR{args.pr}"
    out_dir.mkdir(parents=True, exist_ok=True)
    out_path = out_dir / f"{head_sha}.md"

    # Render header
    date_str = dt.datetime.utcnow().strftime("%Y-%m-%d")
    header = render_header(date_str, args.agent, head_sha, branch, branch_url, args.pr, pr_url, f"{head_sha}.md", repo_full)

    # Render items
    lines = [header]
    checkbox_block = "\n".join([
        "- [ ] Fixed",
        "- [ ] Test Written",
        "- [ ] Suggestion Ignored",
        "",
    ])
    for it in collected:
        lines.append(f"### {it['title']}")
        lines.append("")
        lines.append(checkbox_block)
        lines.append("```text")
        lines.append(it["prompt"])
        lines.append("```")
        lines.append("")
        lines.append("{response}")
        lines.append("")

    # Conclusion
    lines.append("---")
    lines.append("")
    lines.append("## Conclusion")
    lines.append("")
    lines.append("| Accepted | Rejected | Remarks |")
    lines.append("|-----------|----------|---------|")
    lines.append("| {accepted_count} | {rejected_count} | {remarks} |")
    lines.append("")
    lines.append("{comments}")
    lines.append("")

    out_path.write_text("\n".join(lines), encoding="utf-8")
    print(str(out_path))


if __name__ == "__main__":
    main()
