#!/usr/bin/env python3
"""Summarize repeated Go benchmark samples without third-party dependencies.

Usage: python3 summarize.py before.txt after.txt > results.json
"""

import json
import re
import statistics
import sys
from pathlib import Path


def read_samples(path):
    samples = {}
    for line in Path(path).read_text().splitlines():
        if not line.startswith("Benchmark"):
            continue
        fields = line.split()
        name = re.sub(r"-\d+$", "", fields[0])
        metrics = samples.setdefault(name, {})
        for value, unit in zip(fields[2::2], fields[3::2]):
            metrics.setdefault(unit, []).append(float(value))
    return samples


def summary(metrics):
    return {
        unit: {"median": statistics.median(values), "min": min(values),
               "max": max(values), "samples": values}
        for unit, values in metrics.items()
    }


before, after = (read_samples(path) for path in sys.argv[1:3])
if before.keys() != after.keys():
    raise SystemExit("Before/after benchmark names differ")
results = []
for name in before:
    results.append({"name": name, "before": summary(before[name]),
                    "after": summary(after[name])})
json.dump(results, sys.stdout, indent=2)
print()
