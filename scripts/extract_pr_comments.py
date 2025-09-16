#!/usr/bin/env python3
"""
Extract CodeRabbit (or any author) comments from a GitHub Pull Request using gh.

Usage:
  python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit \
      --out docs/audits/coderabbit-pr123-comments.md

Prompts-only mode (extract "Prompt for AI Agents" sections only):
  python3 scripts/extract_pr_comments.py --pr 123 --author coderabbit \
      --prompts-only --prompts-out docs/audits/coderabbit-pr123-prompts.md

If --pr is omitted, the current branch PR is used (via `gh pr view`).
If --out/--prompts-out are omitted, defaults under docs/audits/ are used.

Requires: gh CLI authenticated to the repo.
"""
import argparse
import json
import os
import subprocess
from datetime import datetime

def sh(cmd):
    return subprocess.check_output(cmd, text=True).strip()

def gh_json(args):
    return json.loads(sh(args))

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--pr', type=int, help='Pull request number')
    ap.add_argument('--author', default='coderabbit', help='Substring filter for author login (case-insensitive)')
    ap.add_argument('--out', help='Output markdown file path')
    ap.add_argument('--prompts-only', action='store_true', help='Extract only "Prompt for AI Agents" sections')
    ap.add_argument('--prompts-out', help='Output markdown file for prompts-only mode')
    args = ap.parse_args()

    repo = sh(['gh','repo','view','--json','nameWithOwner','-q','.nameWithOwner'])
    if not args.pr:
        # current branch PR
        try:
            prnum = int(sh(['gh','pr','view','--json','number','-q','.number']))
        except Exception:
            raise SystemExit('No PR number supplied and no PR associated with current branch.')
    else:
        prnum = args.pr

    author_sub = args.author.lower()
    out = args.out or f'docs/audits/{author_sub}-pr{prnum}-comments.md'
    prompts_out = args.prompts_out or f'docs/audits/{author_sub}-pr{prnum}-prompts.md'
    os.makedirs(os.path.dirname(out), exist_ok=True)
    os.makedirs(os.path.dirname(prompts_out), exist_ok=True)

    # Issue (conversation) comments
    issue_comments = gh_json(['gh','api',f'repos/{repo}/issues/{prnum}/comments','--paginate'])
    # Review comments (inline, on diffs)
    review_comments = gh_json(['gh','api',f'repos/{repo}/pulls/{prnum}/comments','--paginate'])
    # Reviews (with bodies)
    reviews = gh_json(['gh','api',f'repos/{repo}/pulls/{prnum}/reviews','--paginate'])

    items = []
    def add_item(kind, c):
        user = (c.get('user') or {}).get('login','')
        if author_sub in user.lower():
            created = c.get('created_at') or c.get('submitted_at') or ''
            body = c.get('body') or ''
            path = c.get('path','')
            line = c.get('line') or c.get('original_line') or ''
            url = c.get('html_url') or c.get('pull_request_url') or ''
            items.append({
                'kind': kind,
                'created': created,
                'user': user,
                'body': body,
                'path': path,
                'line': line,
                'url': url,
            })

    for c in issue_comments:
        add_item('issue_comment', c)
    for c in review_comments:
        add_item('review_comment', c)
    for r in reviews:
        # skip bare approvals without body
        if r.get('body'):
            add_item('review', r)

    # sort by created time
    def parse_dt(s):
        try:
            return datetime.fromisoformat(s.replace('Z','+00:00'))
        except Exception:
            return datetime.min
    items.sort(key=lambda x: parse_dt(x['created']))

    if args.prompts_only:
        import re
        prompts = []
        for it in items:
            body = it['body'] or ''
            # Look for fenced code after the header phrase
            for m in re.finditer(r'(?is)prompt\s*for\s*ai\s*agents[^`]*```(.*?)```', body):
                txt = m.group(1).strip()
                if txt:
                    prompts.append({**it, 'prompt': txt})
            # Also capture inline prompts (until blank line or next header)
            for m in re.finditer(r'(?is)(prompt\s*for\s*ai\s*agents\s*:?[\t ]*)(.+?)(?:\n\s*\n|\n#+|\n>\s|$)', body):
                txt = m.group(2).strip()
                # Skip if already captured from code fence
                if txt and not any(txt in p['prompt'] for p in prompts):
                    prompts.append({**it, 'prompt': txt})

        with open(prompts_out, 'w', encoding='utf-8') as f:
            f.write(f"# CodeRabbit Prompts for PR #{prnum}\n\n")
            f.write(f"Repo: {repo}\n\n")
            for p in prompts:
                header = f"- [{p['kind']}] {p['created']} by {p['user']}"
                loc = f" ({p['path']}:{p['line']})" if p['path'] else ''
                link = f"\n  URL: {p['url']}" if p['url'] else ''
                f.write(header+loc+"\n\n")
                f.write(p['prompt'].rstrip()+"\n\n")
                if link:
                    f.write(link+"\n\n")
        print(prompts_out)
        return

    # Default: full comments dump
    with open(out,'w',encoding='utf-8') as f:
        f.write(f"# CodeRabbit Comments for PR #{prnum}\n\n")
        f.write(f"Repo: {repo}\n\n")
        for it in items:
            header = f"- [{it['kind']}] {it['created']} by {it['user']}"
            loc = f" ({it['path']}:{it['line']})" if it['path'] else ''
            link = f"\n  URL: {it['url']}" if it['url'] else ''
            f.write(header+loc+"\n\n")
            f.write(it['body'].rstrip()+"\n\n")
            if link:
                f.write(link+"\n\n")

    print(out)

if __name__ == '__main__':
    main()
