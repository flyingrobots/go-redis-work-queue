#!/usr/bin/env python3
"""Regression tests for the CodeRabbit worksheet generator."""

from __future__ import annotations

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest


SCRIPT = Path(__file__).with_name("generate_cr_worksheet.py")


class GenerateCRWorksheetTest(unittest.TestCase):
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
                if args == ["api", "repos/flyingrobots/go-redis-work-queue/pulls/42"]:
                    print(json.dumps({
                        "head": {
                            "sha": "abc123",
                            "ref": "feature/test",
                            "repo": {"full_name": "flyingrobots/go-redis-work-queue"},
                        },
                        "html_url": "https://example.test/pull/42",
                    }))
                    raise SystemExit(0)

                if not args or args[0] != "api":
                    raise SystemExit(f"unexpected gh invocation: {args}")

                endpoint = args[1]
                responses = {
                    "repos/flyingrobots/go-redis-work-queue/issues/42/comments": [
                        [{
                            "user": {"login": "coderabbitai"},
                            "created_at": "2026-01-01T00:00:00Z",
                            "body": "Prompt for AI Agents\\n```\\nrepair first page\\n```",
                        }],
                        [{
                            "user": {"login": "coderabbitai"},
                            "created_at": "2026-01-02T00:00:00Z",
                            "body": "Prompt for AI Agents\\n```\\nrepair second page\\n```",
                        }],
                    ],
                    "repos/flyingrobots/go-redis-work-queue/pulls/42/comments": [[{
                        "user": {"login": "coderabbitai"},
                        "created_at": "2026-01-03T00:00:00Z",
                        "body": "Prompt for AI Agents\\n```\\nrepair inline\\n```",
                        "path": "queue.go",
                        "line": 17,
                    }]],
                    "repos/flyingrobots/go-redis-work-queue/pulls/42/reviews": [[{
                        "user": {"login": "someone-else"},
                        "submitted_at": "2026-01-04T00:00:00Z",
                        "body": "Must not be included",
                    }]],
                }
                pages = responses[endpoint]
                if "--slurp" in args:
                    print(json.dumps(pages))
                else:
                    for page in pages:
                        print(json.dumps(page))
                """
            ),
            encoding="utf-8",
        )
        mock.chmod(0o755)

    def test_generates_from_every_paginated_page(self) -> None:
        env = os.environ.copy()
        env["PATH"] = f"{self.bin_dir}{os.pathsep}{env['PATH']}"
        result = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "--repo",
                "flyingrobots/go-redis-work-queue",
                "--pr",
                "42",
                "--out-root",
                "worksheets",
            ],
            cwd=self.root,
            env=env,
            check=False,
            capture_output=True,
            text=True,
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        output = (self.root / "worksheets" / "PR42" / "abc123.md").read_text(
            encoding="utf-8"
        )
        self.assertIn("repair first page", output)
        self.assertIn("repair second page", output)
        self.assertIn("repair inline", output)
        self.assertIn("queue.go:17", output)
        self.assertNotIn("Must not be included", output)


if __name__ == "__main__":
    unittest.main()
