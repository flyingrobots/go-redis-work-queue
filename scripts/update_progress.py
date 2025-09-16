#!/usr/bin/env python3
"""
Recompute weighted overall progress from docs/features-ledger.md and update
the fixed-width progress bars in both the ledger and README.md.

Weight formula per row:
  - Parse Progress % column (e.g., "55% (conf: med)" -> 55.0)
  - Resolve Code links to local paths and sum Go LOC under those paths
  - Weight w = max(0.5, log10(LOC + 10) / 3)
  - Overall = sum(pct * w) / sum(w)

Usage:
  python3 scripts/update_progress.py
"""
import os
import re
import math
from pathlib import Path
from typing import List, Tuple, Dict

ROOT = Path(__file__).resolve().parents[1]
LEDGER = ROOT / "docs" / "features-ledger.md"
README = ROOT / "README.md"

def read_text(p: Path) -> str:
    return p.read_text(encoding="utf-8")

def write_text(p: Path, s: str) -> None:
    p.write_text(s, encoding="utf-8", newline="\n")

def slugify(name: str) -> str:
    s = name.strip().lower()
    s = re.sub(r"[^a-z0-9]+", "-", s)
    return s.strip('-') or "group"

def parse_all_tables(md: str) -> Tuple[List[Dict], Dict[str, List[Dict]]]:
    lines = md.splitlines()
    current_group = None
    table_headers = []  # list of tuples (idx, group_slug)
    for i, line in enumerate(lines):
        if line.startswith("### "):
            current_group = slugify(line[4:])
        s = line.strip()
        if s.startswith("|") and ("Emoji" in s) and ("Progress %" in s):
            table_headers.append((i, current_group or "root"))

    all_rows: List[Dict] = []
    groups: Dict[str, List[Dict]] = {}
    for header_idx, group in table_headers:
        i = header_idx + 2  # skip header row + separator row
        while i < len(lines):
            raw = lines[i]
            if not raw.strip().startswith("|"):
                break
            cells = [c.strip() for c in raw.strip().strip('|').split('|')]
            if len(cells) < 5:
                break
            row = {"cells": cells, "raw": raw, "group": group}
            all_rows.append(row)
            groups.setdefault(group, []).append(row)
            i += 1
    return all_rows, groups

def extract_links(md_cell: str):
    # Find markdown links [text](href)
    links = re.findall(r"\[[^\]]+\]\(([^)]+)\)", md_cell)
    return links

def go_loc(paths):
    total = 0
    for href in paths:
        # Normalize relative links like ../internal/x
        p = (LEDGER.parent / href).resolve()
        # Ensure p is within repo
        try:
            p.relative_to(ROOT)
        except ValueError:
            continue
        if p.is_file():
            if p.suffix == ".go":
                try:
                    with open(p, "r", encoding="utf-8", errors="ignore") as f:
                        total += sum(1 for _ in f)
                except Exception:
                    pass
        elif p.is_dir():
            for dirpath, _, files in os.walk(p):
                for fn in files:
                    if not fn.endswith('.go'):
                        continue
                    fp = Path(dirpath) / fn
                    try:
                        with open(fp, "r", encoding="utf-8", errors="ignore") as f:
                            total += sum(1 for _ in f)
                    except Exception:
                        pass
    return total

def compute_weighted_progress(md: str):
    # Indices based on header columns
    # [ '', ' Emoji ', ' Feature ', ' Area ', ' Spec ', ' Code ', ' Status ', ' Progress % ', ' Bar ', ' Current State ', ... ]
    # After split, leading/trailing blanks included
    # Build normalized header cells to locate column positions (from first header)
    header_line = next((line for line in md.splitlines() if line.strip().startswith("|") and ("Emoji" in line) and ("Progress %" in line)), None)
    if header_line is None:
        raise RuntimeError("Could not find table header for column index resolution")
    header_cells = [c.strip() for c in header_line.strip().strip('|').split('|')]
    # Map indices by header name
    prog_idx = header_cells.index("Progress %")
    code_idx = header_cells.index("Code")
    rows, groups = parse_all_tables(md)

    def row_metrics(row: Dict):
        parts = row["cells"]
        raw = row["raw"]
        prog_cell = parts[prog_idx]
        m = re.search(r"(\d+(?:\.\d+)?)%", prog_cell)
        if not m:
            return 0.0, 0.0, 0
        pct = float(m.group(1))
        code_cell = parts[code_idx]
        links = extract_links(code_cell)
        loc = go_loc(links)
        # optional weight override via HTML comment: <!-- weight: X -->
        override = None
        mo = re.search(r"<!--\s*weight\s*:\s*([0-9.]+)\s*-->", raw)
        if mo:
            try:
                override = float(mo.group(1))
            except Exception:
                override = None
        w = override if override is not None else max(0.5, math.log10(loc + 10) / 3.0)
        return pct, w, loc

    # Overall
    num = 0.0
    den = 0.0
    for r in rows:
        pct, w, _ = row_metrics(r)
        num += pct * w
        den += w
    overall = 0.0 if den == 0 else num / den

    # Per-group
    group_stats = {}
    for g, lst in groups.items():
        gnum = 0.0
        gden = 0.0
        gloc = 0
        for r in lst:
            pct, w, loc = row_metrics(r)
            gnum += pct * w
            gden += w
            gloc += loc
        gprog = 0.0 if gden == 0 else gnum / gden
        group_stats[g] = {"progress": gprog, "weight": gden, "kloc": gloc/1000.0, "count": len(lst)}

    return overall, group_stats

def render_bar(percent: float, width: int = 40) -> str:
    p = max(0.0, min(100.0, percent))
    total = width
    filled = int(p/100.0 * total)
    frac = (p/100.0 * total) - filled
    bar = ""
    bar += "█" * min(filled, total)
    if filled < total:
        bar += "▓" if frac > 0.01 else ""
    # recompute remaining after adding partial
    used = len(bar)
    if used < total:
        bar += "░" * (total - used)
    return f"{bar} {p:.0f}%"

def render_markers(width: int = 40) -> tuple[str,str]:
    """
    Render a fixed-width markers line and labels for MVP, Alpha, Beta, v1.0.0.
    Markers at 1/4, 2/4, 3/4 and end.
    """
    q1 = round(width * 1/4)
    q2 = round(width * 2/4)
    q3 = round(width * 3/4)
    positions = (q1-1, q2-1, q3-1, width-1)
    line = ["-"] * width
    for pos in positions:
        if 0 <= pos < width:
            line[pos] = "|"
    line_str = "".join(line)
    # labels line
    labels = [" "] * width
    def place(text: str, center: int):
        start = max(0, min(width - len(text), center - len(text)//2))
        for i,ch in enumerate(text):
            if 0 <= start+i < width:
                labels[start+i] = ch
    place("MVP", q1-1)
    place("Alpha", q2-1)
    place("Beta", q3-1)
    place("v1.0.0", width-1 - (len("v1.0.0")//2))
    labels_str = "".join(labels)
    return line_str, labels_str

def replace_progress_block(text: str, bar_block: str) -> str:
    begin = "<!-- progress:begin -->"
    end = "<!-- progress:end -->"
    if begin in text and end in text:
        pre, rest = text.split(begin, 1)
        _, post = rest.split(end, 1)
        return pre + begin + "\n" + bar_block + "\n" + end + post
    # If markers missing, append at end
    return text + "\n\n" + begin + "\n" + bar_block + "\n" + end + "\n"

def insert_or_replace_group_block(md: str, group_heading: str, slug: str, bar_block: str) -> str:
    lines = md.splitlines()
    out = []
    i = 0
    inserted = False
    begin = f"<!-- group-progress:{slug}:begin -->"
    end = f"<!-- group-progress:{slug}:end -->"
    while i < len(lines):
        line = lines[i]
        out.append(line)
        if line.startswith("### ") and slugify(line[4:]) == slug:
            # search for existing block right after
            j = i + 1
            if j < len(lines) and begin in lines[j]:
                # Replace existing block
                out.append(begin)
                out.append(bar_block)
                # skip until end marker
                j += 1
                while j < len(lines) and end not in lines[j]:
                    j += 1
                if j < len(lines):
                    # append end and advance past it
                    out.append(end)
                    i = j  # loop will i+=1
                inserted = True
            else:
                # Insert new block right after heading
                out.append(begin)
                out.append(bar_block)
                out.append(end)
                inserted = True
        i += 1
    return "\n".join(out) + ("\n" if not md.endswith("\n") else "")

def main():
    ledger_md = read_text(LEDGER)
    overall, group_stats = compute_weighted_progress(ledger_md)
    width = 40
    bar = render_bar(overall, width)
    markers, labels = render_markers(width)
    bar_block = "```text\n" + bar + "\n" + markers + "\n" + labels + "\n" + "```"

    # Update overall in ledger
    new_ledger = replace_progress_block(ledger_md, bar_block)
    # Update per-group blocks beneath each group heading
    for g, stats in group_stats.items():
        gbar = render_bar(stats["progress"], width)
        gmarkers, glabels = render_markers(width)
        extra = f"weight={stats['weight']:.2f} features={stats['count']} kloc={stats['kloc']:.1f}"
        gblock = "```text\n" + gbar + "\n" + gmarkers + "\n" + glabels + "\n" + extra + "\n" + "```"
        # Find the original group heading line text (we only need slug for placement)
        new_ledger = insert_or_replace_group_block(new_ledger, g, g, gblock)
    # After progress blocks, ensure KLoC column exists and is populated
    new_ledger = add_or_update_kloc_column(new_ledger)
    write_text(LEDGER, new_ledger)

    readme_md = read_text(README)
    new_readme = replace_progress_block(readme_md, bar_block)
    write_text(README, new_readme)

def add_or_update_kloc_column(md: str) -> str:
    lines = md.splitlines()
    i = 0
    while i < len(lines):
        line = lines[i]
        s = line.strip()
        if s.startswith('|') and ("Emoji" in s) and ("Progress %" in s):
            header_cells = [c.strip() for c in s.strip('|').split('|')]
            # identify indices
            try:
                code_idx = header_cells.index("Code")
            except ValueError:
                i += 1
                continue
            # Prepare header used for rows; may be extended with KLoC
            header_for_rows = header_cells[:]
            if any(c.lower().startswith("kloc") for c in header_cells):
                # find existing kloc index
                for idx, c in enumerate(header_cells):
                    if c.lower().startswith("kloc"):
                        kloc_idx = idx
                        break
                new_header_cells = header_cells
            else:
                # insert KLoC after Code
                kloc_idx = code_idx + 1
                new_header_cells = header_cells[:kloc_idx] + ["KLoC (approx)"] + header_cells[kloc_idx:]
                lines[i] = "|" + " | ".join(new_header_cells) + " |"
                # replace separator row beneath
                if i + 1 < len(lines) and lines[i+1].strip().startswith('|'):
                    sep = "|" + "|".join([" --- "] * len(new_header_cells)) + "|"
                    lines[i+1] = sep
                header_for_rows = new_header_cells
            # Resolve column indices for rows
            def idx(name: str):
                try:
                    return header_for_rows.index(name)
                except ValueError:
                    return None
            code_i = idx("Code")
            status_i = idx("Status")
            emoji_i = idx("Emoji")
            # walk rows and update kloc cell
            j = i + 2
            while j < len(lines) and lines[j].strip().startswith("|"):
                row_cells = [c.strip() for c in lines[j].strip().strip('|').split('|')]
                # pad cells if needed
                while len(row_cells) < len(header_for_rows):
                    row_cells.insert(kloc_idx, "")
                # compute kloc from code cell
                code_cell = row_cells[code_i] if code_i is not None and code_i < len(row_cells) else ""
                links = extract_links(code_cell)
                loc = go_loc(links)
                row_cells[kloc_idx] = f"{loc/1000.0:.1f}"
                # Set status emoji based on Status column
                if status_i is not None and emoji_i is not None and status_i < len(row_cells) and emoji_i < len(row_cells):
                    status = row_cells[status_i].lower()
                    if status.startswith('planned'):
                        emoji = '📋'
                    elif status.startswith('in progress'):
                        emoji = '⏳'
                    elif status.startswith('mvp'):
                        emoji = '🚼'
                    elif status.startswith('alpha'):
                        emoji = '🅰️'
                    elif status.startswith('beta'):
                        emoji = '🅱️'
                    elif status.startswith('v1') or 'v1' in status or 'shipped' in status:
                        emoji = '✅'
                    else:
                        emoji = '⏳'
                    row_cells[emoji_i] = emoji
                lines[j] = "|" + " | ".join(row_cells) + " |"
                j += 1
            i = j
        else:
            i += 1
    return "\n".join(lines) + ("\n" if not md.endswith("\n") else "")

if __name__ == "__main__":
    main()
