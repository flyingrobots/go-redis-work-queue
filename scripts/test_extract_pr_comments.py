#!/usr/bin/env python3
"""Regression tests for the documented PR comment extractor."""

from __future__ import annotations

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest


SCRIPT = Path(__file__).with_name("extract_pr_comments.py")


class ExtractPRCommentsTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp_dir.cleanup)
        self.root = Path(self.temp_dir.name)
        self.bin_dir = self.root / "bin"
        self.bin_dir.mkdir()
        self._write_mock_gh()

    def _write_mock_gh(self) -> None:
        mock = self.bin_dir / "gh"
        mock.write_text(
            textwrap.dedent(
                """\
                #!/usr/bin/env python3
                import json
                import sys

                args = sys.argv[1:]
                if args[:2] == ["repo", "view"]:
                    print("flyingrobots/go-redis-work-queue")
                    raise SystemExit(0)

                if not args or args[0] != "api":
                    raise SystemExit(f"unexpected gh invocation: {args}")

                endpoint = args[1]
                responses = {
                    "repos/flyingrobots/go-redis-work-queue/issues/42/comments": [
                        [{
                            "user": {"login": "coderabbitai"},
                            "created_at": "2026-01-01T00:00:00Z",
                            "body": "First issue comment\\n\\nPrompt for AI Agents\\n```\\nrepair issue\\n```",
                            "html_url": "https://example.test/issue",
                        }],
                        [{
                            "user": {"login": "someone-else"},
                            "created_at": "2026-01-02T00:00:00Z",
                            "body": "Must not be included",
                        }],
                    ],
                    "repos/flyingrobots/go-redis-work-queue/pulls/42/comments": [[{
                        "user": {"login": "coderabbitai"},
                        "created_at": "2026-01-03T00:00:00Z",
                        "body": "Inline review comment",
                        "path": "queue.go",
                        "line": 17,
                        "html_url": "https://example.test/inline",
                    }]],
                    "repos/flyingrobots/go-redis-work-queue/pulls/42/reviews": [[{
                        "user": {"login": "coderabbitai"},
                        "submitted_at": "2026-01-04T00:00:00Z",
                        "body": "Review summary",
                        "html_url": "https://example.test/review",
                    }]],
                }
                print(json.dumps(responses[endpoint]))
                """
            ),
            encoding="utf-8",
        )
        mock.chmod(0o755)

    def run_script(self, *args: str) -> subprocess.CompletedProcess[str]:
        env = os.environ.copy()
        env["PATH"] = f"{self.bin_dir}{os.pathsep}{env['PATH']}"
        return subprocess.run(
            [sys.executable, str(SCRIPT), *args],
            cwd=self.root,
            env=env,
            check=False,
            capture_output=True,
            text=True,
        )

    def test_extracts_every_paginated_comment_type(self) -> None:
        result = self.run_script(
            "--pr",
            "42",
            "--author",
            "coderabbit",
            "--out",
            "comments.md",
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        output = (self.root / "comments.md").read_text(encoding="utf-8")
        self.assertIn("Review Comments Matching 'coderabbit'", output)
        self.assertIn("First issue comment", output)
        self.assertIn("Inline review comment", output)
        self.assertIn("queue.go:17", output)
        self.assertIn("Review summary", output)
        self.assertNotIn("Must not be included", output)

    def test_prompts_only_extracts_fenced_prompt(self) -> None:
        result = self.run_script(
            "--pr",
            "42",
            "--prompts-only",
            "--prompts-out",
            "prompts.md",
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        output = (self.root / "prompts.md").read_text(encoding="utf-8")
        self.assertIn("Review Prompts Matching 'coderabbit'", output)
        self.assertIn("repair issue", output)
        self.assertNotIn("Inline review comment", output)


if __name__ == "__main__":
    unittest.main()
