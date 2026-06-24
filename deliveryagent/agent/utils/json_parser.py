"""
Robust JSON extraction from LLM text output.

LLM frequently wraps JSON in markdown code blocks, adds explanatory text,
or produces minor formatting issues (trailing commas, single quotes, unquoted keys).
This module provides a unified fallback chain.

Usage:
    data = extract_json(text)         # dict | list | None
    objs = extract_all_json(text)     # list of parsed objects (for multi-JSON)
"""
from __future__ import annotations

import json
import re
from typing import Any


def extract_json(text: str) -> Any | None:
    """Extract the first valid JSON value from LLM output text.

    Fallback chain:
      1. Direct json.loads(text)
      2. Strip markdown code blocks → loads
      3. Strip extra text, extract outermost {} or [] → loads
      4. Fix common issues (trailing commas, single quotes, unquoted keys) → loads

    Returns:
        Parsed JSON (dict | list) or None if all attempts fail.
    """
    if not text:
        return None
    text = text.strip()
    if not text:
        return None

    for candidate in _collect_candidates(text):
        result = _try_parse(candidate)
        if result is not None:
            return result

    return None


def extract_all_json(text: str, default: list | None = None) -> list[Any]:
    """Extract all valid JSON objects from text (for multi-tool-call output).

    Falls back to *default* (empty list) when nothing parseable is found.
    """
    result: list[Any] = default if default is not None else []
    if not text:
        return result
    text = text.strip()
    if not text:
        return result

    # If whole text parses as a JSON array, return its items.
    single = _try_parse(text)
    if isinstance(single, list):
        return single

    # Walk through text and extract every outermost {…} block.
    pos = 0
    while True:
        obj = _extract_outermost(text, pos)
        if obj is None:
            break
        parsed = _try_parse(obj)
        if parsed is not None:
            result.append(parsed)
        idx = text.index(obj, pos)
        pos = idx + len(obj)

    return result


# ─── Internal helpers ─────────────────────────────────


def _collect_candidates(text: str) -> list[str]:
    """Build progressively cleaned text candidates."""
    candidates: list[str] = []
    seen: set[str] = set()

    def add(s: str) -> None:
        s = s.strip()
        if s and s not in seen:
            seen.add(s)
            candidates.append(s)

    add(text)

    # 1. Content inside markdown code blocks
    inner = _strip_code_block(text)
    if inner != text:
        add(inner)

    # 2. Outermost {} or [] from raw text
    for wrapper in ("{}", "[]"):
        obj = _extract_outermost_at(text, wrapper)
        if obj is not None:
            add(obj)

    # 3. Outermost from code-stripped text (if different)
    if inner != text:
        for wrapper in ("{}", "[]"):
            obj = _extract_outermost_at(inner, wrapper)
            if obj is not None:
                add(obj)

    return candidates


def _try_parse(s: str) -> Any | None:
    """Try direct parse followed by one round of fix-and-retry."""
    try:
        return json.loads(s)
    except json.JSONDecodeError:
        pass

    fixed = _fix_common(s)
    if fixed != s:
        try:
            return json.loads(fixed)
        except json.JSONDecodeError:
            pass

    return None


# ─── Text cleaning ──────────────────────────────────

_CODE_BLOCK_RE = re.compile(r"```(?:json)?\s*\n?([\s\S]*?)\n?```")


def _strip_code_block(text: str) -> str:
    """Return content inside the first markdown code block, or original text."""
    m = _CODE_BLOCK_RE.search(text)
    return m.group(1).strip() if m else text


def _extract_outermost(text: str, pos: int = 0) -> str | None:
    """Extract outermost balanced pair ({} or []) starting from *pos*."""
    return _extract_outermost_at(text, "{}", pos) or _extract_outermost_at(text, "[]", pos)


def _extract_outermost_at(text: str, wrapper: str, pos: int = 0) -> str | None:
    """Extract outermost balanced *wrapper* (e.g. '{}') from *pos*."""
    open_ch, close_ch = wrapper[0], wrapper[1]
    depth = 0
    start = -1
    for i in range(pos, len(text)):
        ch = text[i]
        if ch == open_ch:
            if depth == 0:
                start = i
            depth += 1
        elif ch == close_ch:
            depth -= 1
            if depth == 0 and start >= 0:
                return text[start : i + 1]
    return None


_UNQUOTED_KEY_RE = re.compile(r"([{,]\s*)(\w+)(\s*:)")


def _fix_common(text: str) -> str:
    """Repair common JSON formatting issues from LLM output.

    Safe for LLM output because:
    - Trailing commas are never valid in JSON.
    - Single quotes: LLM JSON output uses them consistently (all keys and
      string values), so a blanket replacement doesn't break mixed quoting.
    - Unquoted keys: only matched at positions where json.loads already
      failed, so the heuristic is low-risk.
    """
    # Remove trailing commas before ] or }
    text = re.sub(r",\s*([}\]])", r"\1", text)

    # Single quotes → double quotes (safe for LLM-generated JSON)
    if "'" in text and '"' not in text:
        text = text.replace("'", '"')

    # Quote unquoted keys: {key: "val"} → {"key": "val"}
    text = _UNQUOTED_KEY_RE.sub(r'\1"\2"\3', text)

    return text
