#!/usr/bin/env python3
import subprocess
import sys
from pathlib import Path

def main() -> int:
    result = subprocess.run(["git", "ls-files", "*.yml", "*.yaml"], capture_output=True, text=True, check=True)
    files = [line.strip() for line in result.stdout.splitlines() if line.strip()]
    missing: list[str] = []
    for name in files:
        path = Path(name)
        if not path.exists():
            continue
        data = path.read_bytes()
        if data and not data.endswith(b"\n"):
            missing.append(name)
    if missing:
        for name in missing:
            print(f"YAML file missing trailing newline: {name}", file=sys.stderr)
        return 1
    return 0

if __name__ == "__main__":
    sys.exit(main())
