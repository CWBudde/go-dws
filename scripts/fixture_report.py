#!/usr/bin/env python3
"""Honest DWScript fixture compatibility report.

Runs every testdata/fixtures/*/*.pas through ./bin/dwscript run and compares the
normalized output to the sibling .txt. Prints a per-category pass/fail table and a
total. This is the *ground-truth* compatibility metric for the port -- unlike the
in-repo Go harness, it does not skip categories.

Usage:
    go build -o bin/dwscript ./cmd/dwscript
    python3 scripts/fixture_report.py [--category NAME] [--list-fails] [--timeout SECS]
"""
import argparse
import collections
import os
import subprocess
import sys

BASE = "testdata/fixtures"
CLI = "./bin/dwscript"


def norm(s: str) -> str:
    lines = [ln.rstrip() for ln in s.replace("\r\n", "\n").split("\n")]
    return "\n".join(lines).strip()


def run_one(pf: str, timeout: int) -> str:
    try:
        r = subprocess.run([CLI, "run", pf], capture_output=True, timeout=timeout)
        return (r.stdout + r.stderr).decode("utf-8", errors="replace")
    except subprocess.TimeoutExpired:
        return "__TIMEOUT__"


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--category", help="only run this category")
    ap.add_argument("--list-fails", action="store_true", help="print failing filenames")
    ap.add_argument("--timeout", type=int, default=20)
    args = ap.parse_args()

    if not os.path.exists(CLI):
        print(f"error: {CLI} not found. Build it: go build -o bin/dwscript ./cmd/dwscript",
              file=sys.stderr)
        return 2

    cats = collections.OrderedDict()
    fails = []
    for cat in sorted(os.listdir(BASE)):
        if args.category and cat != args.category:
            continue
        d = os.path.join(BASE, cat)
        if not os.path.isdir(d):
            continue
        pas = sorted(f for f in os.listdir(d) if f.endswith(".pas"))
        if not pas:
            continue
        p = f = noexp = 0
        for name in pas:
            pf = os.path.join(d, name)
            tf = pf[:-4] + ".txt"
            if not os.path.exists(tf):
                noexp += 1
                continue
            exp = norm(open(tf, encoding="utf-8", errors="replace").read())
            got = norm(run_one(pf, args.timeout))
            if got == exp:
                p += 1
            else:
                f += 1
                fails.append(f"{cat}/{name}")
        cats[cat] = (len(pas), p, f, noexp)

    tp = tf = tno = ttot = 0
    print(f"{'Category':<26}{'Tot':>5}{'Pass':>6}{'Fail':>6}{'NoExp':>6}{'Pass%':>7}")
    for c, (t, p, f, ne) in cats.items():
        scored = p + f
        pct = (100 * p / scored) if scored else 0
        print(f"{c:<26}{t:>5}{p:>6}{f:>6}{ne:>6}{pct:>6.0f}%")
        tp += p
        tf += f
        tno += ne
        ttot += t
    scored = tp + tf
    print("-" * 56)
    print(f"{'TOTAL':<26}{ttot:>5}{tp:>6}{tf:>6}{tno:>6}"
          f"{(100 * tp / scored) if scored else 0:>6.0f}%")

    if args.list_fails:
        print("\nFailing fixtures:")
        for name in fails:
            print(f"  {name}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
