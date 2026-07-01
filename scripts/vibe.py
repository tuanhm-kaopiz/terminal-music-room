#!/usr/bin/env python3
"""Vibe DevKit validator — pipeline artifacts, gates, and tooling."""
from __future__ import annotations

import argparse
import json
import re
import sys
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

PHASES = [
    ("idea", "idea.md", "G0", "Idea captured"),
    ("clarify", "clarify.md", "G1", "Clarify complete"),
    ("spec", "spec.md", "G2", "Spec approved"),
    ("architecture", "architecture.md", "G3", "Architecture approved"),
    ("tasks", "tasks.md", "G4", "Tasks ready"),
    ("code", "tasks.md", "G5", "Code complete"),
    ("review", "review.md", "G6", "Ship ready"),
]

CHECKBOX_RE = re.compile(r"^- \[([ xX])\]", re.M)
OPEN_Q_BLOCKING_RE = re.compile(
    r"## Open questions \(blocking\).*?(?=##|\Z)", re.S | re.I
)
TASK_DONE_RE = re.compile(r"^- \[([ xX])\].*\*\*T-\d+", re.M)
TASK_MAP_RE = re.compile(r"Maps to:\s*(.+)", re.I)
AC_ID_RE = re.compile(r"\bAC-\d+\b", re.I)
TASK_ID_RE = re.compile(r"\*\*T-\d+:\*\*", re.I)
ADR_RE = re.compile(r"### ADR-\d+:", re.I)
GIVEN_WHEN_THEN_RE = re.compile(
    r"Given\s+.+\s+When\s+.+\s+Then\s+", re.I | re.S
)
STATUS_APPROVED_RE = re.compile(r"\*\*Status:\*\*\s*approved\b", re.I)

CODE_INTENT_PATTERNS = [
    r"\bimplement\b",
    r"\bbuild\b",
    r"\bcode\s+it\b",
    r"\bwrite\s+(the\s+)?code\b",
    r"\bfix\s+(the\s+)?bug\b",
    r"\badd\s+(the\s+)?(feature|endpoint|api|route|component)\b",
    r"\bcreate\s+(the\s+)?(feature|endpoint|api|route|component)\b",
    r"\brefactor\b.*\b(now|ngay)\b",
    r"\btạo\s+code\b",
    r"\blàm\s+luôn\b",
    r"\bimplement\s+ngay\b",
    r"\bjust\s+code\b",
    r"\bvibe\s+coding\b.*\b(now|luôn|ngay)\b",
    r"\bthêm\s+tính\s+năng\b",
    r"\bviết\s+code\b",
    r"\bsửa\s+bug\b",
    r"\bship\s+it\b",
    r"\bmerge\s+and\s+deploy\b",
]

OVERRIDE_PATTERNS = [
    r"GATE\s*OVERRIDE",
    r"QUICK:\s*yes",
    r"/vibe-",
    r"\bvibe-(idea|clarify|spec|architecture|tasks|code|review|status)\b",
]


@dataclass
class VibeConfig:
    artifacts_dir: str = "docs/vibe"
    active_feature: str | None = None
    gate_mode: str = "strict"
    quick_change_loc_threshold: int = 10
    require_human_approval: bool = False
    language: str = "vi"
    commands: dict[str, str] = field(default_factory=dict)

    @classmethod
    def load(cls, project_root: Path) -> VibeConfig:
        config_path = project_root / "vibe.config.yaml"
        cfg = cls()
        if not config_path.exists():
            return cfg
        text = config_path.read_text(encoding="utf-8", errors="replace")

        def scalar(key: str, default: str | None = None) -> str | None:
            m = re.search(rf"^{re.escape(key)}:\s*(.+)$", text, re.M)
            if not m:
                return default
            val = m.group(1).strip().strip("'\"")
            if val in ("null", "~", ""):
                return None
            return val

        artifacts = scalar("artifacts_dir")
        if artifacts:
            cfg.artifacts_dir = artifacts
        active = scalar("active_feature")
        if active:
            cfg.active_feature = active
        mode = scalar("gate_mode")
        if mode:
            cfg.gate_mode = mode.lower()
        lang = scalar("language")
        if lang:
            cfg.language = lang
        threshold = scalar("quick_change_loc_threshold")
        if threshold and threshold.isdigit():
            cfg.quick_change_loc_threshold = int(threshold)
        approval = scalar("require_human_approval")
        if approval:
            cfg.require_human_approval = approval.lower() in ("true", "yes", "1")

        for cmd_key in ("test", "lint", "typecheck"):
            m = re.search(rf"^\s+{cmd_key}:\s*(.+)$", text, re.M)
            if m:
                cfg.commands[cmd_key] = m.group(1).strip().strip("'\"")

        return cfg

    def save_active_feature(self, project_root: Path, feature_ref: str) -> None:
        config_path = project_root / "vibe.config.yaml"
        text = (
            config_path.read_text(encoding="utf-8", errors="replace")
            if config_path.exists()
            else ""
        )
        line = f"active_feature: {feature_ref}"
        if re.search(r"^active_feature:\s*", text, re.M):
            text = re.sub(r"^active_feature:\s*.*$", line, text, flags=re.M)
        else:
            text = text.rstrip() + f"\n\n# Active feature for gate-check and /vibe-* commands\n{line}\n"
        config_path.write_text(text, encoding="utf-8")
        self.active_feature = feature_ref


def gate_checklist_status(content: str, gate_id: str) -> str:
    pattern = re.compile(
        rf"## Gate {re.escape(gate_id)}.*?(?=## |\Z)", re.S | re.I
    )
    section = pattern.search(content)
    if not section:
        return "pending"
    boxes = CHECKBOX_RE.findall(section.group(0))
    if not boxes:
        return "pending"
    if all(b.lower() == "x" for b in boxes):
        return "pass"
    if any(b.lower() == "x" for b in boxes):
        return "in_progress"
    return "pending"


def check_clarify_blocking(content: str) -> bool:
    m = OPEN_Q_BLOCKING_RE.search(content)
    if not m:
        return False
    body = m.group(0)
    lines = [
        ln
        for ln in body.splitlines()
        if ln.strip().startswith("|")
        and not re.match(r"^\|[\s\-#:|]+\|$", ln.strip())
    ]
    for ln in lines:
        cells = [c.strip().lower() for c in ln.strip("|").split("|")]
        if not any(cells):
            continue
        if cells[0] in ("#", "question", "no.", "id"):
            continue
        if all(c in ("", "-", "none", "n/a", "(none)") for c in cells):
            continue
        if any("yes" in c for c in cells if "blocking" not in c):
            return True
        # Row with real question text in first column
        if cells[0] and cells[0] not in ("none", "(none)", "n/a"):
            return True
    return False


def check_code_phase(tasks_content: str) -> str:
    tasks = TASK_DONE_RE.findall(tasks_content)
    if not tasks:
        return "pending"
    if all(t.lower() == "x" for t in tasks):
        return "pass"
    if any(t.lower() == "x" for t in tasks):
        return "in_progress"
    return "pending"


def semantic_issues(phase_key: str, content: str, cfg: VibeConfig) -> list[str]:
    issues: list[str] = []
    if phase_key == "idea":
        ps = re.search(r"## Problem statement\s*\n(.*?)(?=##|\Z)", content, re.S | re.I)
        if ps and len(re.sub(r"\s+", "", ps.group(1))) < 20:
            issues.append("G0: Problem statement section is empty or too short")
        if "## Out of scope" in content:
            oos = re.search(r"## Out of scope\s*\n(.*?)(?=##|\Z)", content, re.S | re.I)
            if oos and not re.search(r"^-\s+\S", oos.group(1), re.M):
                issues.append("G0: Out of scope needs at least one item")
    elif phase_key == "clarify":
        if not re.search(r"## Actors / users", content, re.I):
            issues.append("G1: Actors section missing")
    elif phase_key == "spec":
        ac_ids = AC_ID_RE.findall(content)
        if not ac_ids:
            issues.append("G2: No AC-xxx acceptance criteria found")
        elif cfg.gate_mode == "strict":
            if not GIVEN_WHEN_THEN_RE.search(content) and not re.search(
                r"AC-\d+.*(?:Given|must|should|shall|verify|expect)", content, re.I | re.S
            ):
                issues.append(
                    "G2: AC should use Given/When/Then or measurable wording"
                )
        if cfg.require_human_approval and not STATUS_APPROVED_RE.search(content):
            issues.append("G2: spec.md status must be 'approved' (human sign-off)")
    elif phase_key == "architecture":
        if not ADR_RE.search(content):
            issues.append("G3: At least one ADR-xxx section required")
        if cfg.require_human_approval and not STATUS_APPROVED_RE.search(content):
            issues.append("G3: architecture.md status must be 'approved'")
    elif phase_key == "tasks":
        if not TASK_ID_RE.search(content):
            issues.append("G4: No T-xxx tasks found")
        for line in content.splitlines():
            if "**T-" in line and "Maps to:" not in content:
                issues.append("G4: Tasks should include 'Maps to: AC-xxx'")
                break
    return issues


def list_features(project_root: Path, cfg: VibeConfig | None = None) -> list[Path]:
    cfg = cfg or VibeConfig.load(project_root)
    artifacts = project_root / cfg.artifacts_dir
    if not artifacts.exists():
        return []
    return sorted(
        [d for d in artifacts.iterdir() if d.is_dir() and not d.name.startswith(".")],
        key=lambda p: p.name,
    )


def resolve_feature(
    project_root: Path, feature_arg: str | None = None, cfg: VibeConfig | None = None
) -> Path | None:
    cfg = cfg or VibeConfig.load(project_root)
    features = list_features(project_root, cfg)
    if not features:
        return None

    if feature_arg:
        matches = [f for f in features if feature_arg in f.name]
        return matches[0] if matches else None

    if cfg.active_feature:
        ref = cfg.active_feature.strip().rstrip("/")
        if "/" in ref:
            candidate = project_root / ref
            if candidate.is_dir():
                return candidate
        matches = [f for f in features if ref in f.name]
        if matches:
            return matches[0]

    return features[-1]


def next_feature_number(project_root: Path, cfg: VibeConfig) -> str:
    features = list_features(project_root, cfg)
    max_num = -1
    for f in features:
        m = re.match(r"^(\d{3})-", f.name)
        if m:
            max_num = max(max_num, int(m.group(1)))
    return f"{max_num + 1:03d}"


def slugify(name: str) -> str:
    slug = name.lower().strip()
    slug = re.sub(r"[^a-z0-9]+", "-", slug)
    slug = slug.strip("-")
    return slug or "feature"


def render_template(template_path: Path, replacements: dict[str, str]) -> str:
    text = template_path.read_text(encoding="utf-8")
    for key, val in replacements.items():
        text = text.replace(f"{{{key}}}", val)
    return text


def validate_feature(feature_dir: Path, cfg: VibeConfig) -> dict[str, Any]:
    slug = feature_dir.name
    result: dict[str, Any] = {
        "slug": slug,
        "path": str(feature_dir),
        "phases": [],
        "blockers": [],
        "warnings": [],
    }

    for phase_key, filename, gate_id, gate_name in PHASES:
        fpath = feature_dir / filename
        if not fpath.exists():
            result["phases"].append(
                {
                    "phase": phase_key,
                    "gate": gate_id,
                    "status": "pending",
                    "file": filename,
                }
            )
            if cfg.gate_mode != "relaxed" or phase_key in ("tasks", "code"):
                result["blockers"].append(f"{gate_id}: missing {filename}")
            continue

        content = fpath.read_text(encoding="utf-8", errors="replace")

        if phase_key == "clarify" and check_clarify_blocking(content):
            status = "fail"
            result["blockers"].append(f"{gate_id}: blocking open questions in clarify.md")
        elif phase_key == "code":
            status = check_code_phase(content)
            if status != "pass" and cfg.gate_mode != "relaxed":
                result["blockers"].append(f"{gate_id}: tasks not all complete")
        else:
            status = gate_checklist_status(content, gate_id)

        sem = semantic_issues(phase_key, content, cfg)
        for issue in sem:
            if cfg.gate_mode == "strict":
                result["blockers"].append(issue)
            else:
                result["warnings"].append(issue)

        if status == "pending" and phase_key != "code":
            if cfg.gate_mode == "strict":
                result["blockers"].append(
                    f"{gate_id}: gate checklist incomplete in {filename}"
                )
            elif cfg.gate_mode == "normal" and phase_key in (
                "spec",
                "architecture",
                "tasks",
            ):
                result["warnings"].append(
                    f"{gate_id}: gate checklist incomplete in {filename}"
                )
        elif status == "in_progress":
            if cfg.gate_mode == "strict":
                result["blockers"].append(f"{gate_id}: in progress ({filename})")
            else:
                result["warnings"].append(f"{gate_id}: in progress ({filename})")

        result["phases"].append(
            {
                "phase": phase_key,
                "gate": gate_id,
                "name": gate_name,
                "status": status,
                "file": filename,
            }
        )

    current = "done"
    next_cmd = None
    for p in result["phases"]:
        if p["status"] not in ("pass",):
            current = p["phase"]
            cmd_map = {
                "idea": "/vibe-clarify",
                "clarify": "/vibe-spec",
                "spec": "/vibe-architecture",
                "architecture": "/vibe-tasks",
                "tasks": "/vibe-code",
                "code": "/vibe-review",
                "review": "/vibe-status",
            }
            next_cmd = f"{cmd_map.get(p['phase'], '/vibe-status')} {feature_dir}"
            break
    else:
        next_cmd = f"/vibe-status {feature_dir}"

    result["current_phase"] = current
    result["next_command"] = next_cmd
    result["ship_ready"] = current == "done" or (
        current == "review" and result["phases"][-1]["status"] == "pass"
    )
    return result


def log_gate_override(project_root: Path, prompt: str, reason: str) -> None:
    log_dir = project_root / ".vibe"
    log_dir.mkdir(parents=True, exist_ok=True)
    log_path = log_dir / "override-log.md"
    ts = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    snippet = prompt.replace("\n", " ")[:200]
    entry = f"\n## {ts}\n- **Reason:** {reason}\n- **Prompt:** {snippet}\n"
    if log_path.exists():
        log_path.write_text(log_path.read_text(encoding="utf-8") + entry, encoding="utf-8")
    else:
        header = "# Gate Override Audit Log\n\n> Auto-appended when GATE OVERRIDE or QUICK: yes is used.\n"
        log_path.write_text(header + entry, encoding="utf-8")


def cmd_gate_check(project_root: Path, intent: str) -> dict[str, Any]:
    cfg = VibeConfig.load(project_root)

    for pat in OVERRIDE_PATTERNS:
        if re.search(pat, intent, re.I):
            if re.search(r"GATE\s*OVERRIDE", intent, re.I):
                m = re.search(r"GATE\s*OVERRIDE:\s*(.+)", intent, re.I)
                reason = m.group(1).strip() if m else "unspecified"
                log_gate_override(project_root, intent, reason)
            elif re.search(r"QUICK:\s*yes", intent, re.I):
                log_gate_override(project_root, intent, "QUICK: yes")
            return {"permission": "allow"}

    is_code_intent = any(re.search(p, intent, re.I) for p in CODE_INTENT_PATTERNS)
    if not is_code_intent:
        return {"permission": "allow"}

    features = list_features(project_root, cfg)
    if not features:
        return {
            "permission": "deny",
            "user_message": "Vibe DevKit: Chưa có feature nào. Chạy `vibe new <slug>` hoặc /vibe-idea trước khi code.",
            "agent_message": (
                "GATE FAIL: No features under artifacts_dir. "
                "Run `vibe new <slug>` or /vibe-idea first, "
                "or include GATE OVERRIDE: <reason>."
            ),
        }

    target = resolve_feature(project_root, None, cfg)
    if not target:
        return {"permission": "deny", "user_message": "Vibe DevKit: Không resolve được active feature."}

    if cfg.gate_mode == "relaxed":
        tasks = target / "tasks.md"
        if tasks.exists():
            return {"permission": "allow"}
        return {
            "permission": "deny",
            "user_message": f"Vibe DevKit: Chưa có tasks.md cho {target.name}. Chạy /vibe-tasks.",
            "agent_message": f"GATE 4 FAIL: tasks.md missing at {tasks}.",
        }

    r = validate_feature(target, cfg)
    tasks_phase = next((p for p in r["phases"] if p["phase"] == "tasks"), None)
    if not (target / "tasks.md").exists():
        return {
            "permission": "deny",
            "user_message": (
                f"Vibe DevKit: Chưa có tasks.md cho {target.name}. Chạy /vibe-tasks trước."
            ),
            "agent_message": (
                f"GATE 4 FAIL: tasks.md missing. Active feature: {target.name}. "
                f"Set active_feature in vibe.config.yaml or complete pipeline."
            ),
        }

    if tasks_phase and tasks_phase["status"] != "pass":
        if cfg.gate_mode == "normal" and tasks_phase["status"] == "in_progress":
            return {
                "permission": "allow",
                "warning": "GATE 4 in progress — complete tasks when done.",
            }
        return {
            "permission": "deny",
            "user_message": (
                f"Vibe DevKit: GATE 4 chưa pass ({target.name}). Chạy: {r['next_command']}"
            ),
            "agent_message": (
                f"GATE 4 FAIL for {target.name}. Blockers: {r['blockers']}. "
                "Complete /vibe-tasks or use GATE OVERRIDE."
            ),
        }

    return {"permission": "allow"}


def cmd_new(project_root: Path, slug: str, title: str | None) -> int:
    cfg = VibeConfig.load(project_root)
    artifacts = project_root / cfg.artifacts_dir
    artifacts.mkdir(parents=True, exist_ok=True)

    clean_slug = slugify(slug)
    nnn = next_feature_number(project_root, cfg)
    folder_name = f"{nnn}-{clean_slug}"
    feature_dir = artifacts / folder_name
    if feature_dir.exists():
        print(f"ERROR: Feature folder already exists: {feature_dir}")
        return 1

    feature_dir.mkdir(parents=True)
    devkit_root = Path(__file__).resolve().parent.parent
    template_path = devkit_root / "templates" / "idea.md"
    if not template_path.exists():
        template_path = project_root / "templates" / "idea.md"

    display_name = title or clean_slug.replace("-", " ").title()
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    idea_content = render_template(
        template_path,
        {
            "feature-name": display_name,
            "slug": clean_slug,
            "date": today,
        },
    )
    gate_section = "\n## Gate G0 checklist\n\n- [ ] Problem statement clear\n- [ ] Success metric defined\n- [ ] Out of scope listed\n- [ ] Feature folder created\n"
    if "## Gate G0" not in idea_content:
        idea_content = idea_content.rstrip() + gate_section

    (feature_dir / "idea.md").write_text(idea_content, encoding="utf-8")
    rel_ref = f"{cfg.artifacts_dir}/{folder_name}"
    cfg.save_active_feature(project_root, rel_ref)

    print(f"Created: {feature_dir}")
    print(f"Active feature set: {rel_ref}")
    print(f"Next: /vibe-clarify {feature_dir}")
    return 0


def cmd_set_active(project_root: Path, feature_ref: str) -> int:
    cfg = VibeConfig.load(project_root)
    resolved = resolve_feature(project_root, feature_ref, cfg)
    if not resolved:
        print(f"ERROR: Feature not found: {feature_ref}")
        return 1
    rel = resolved.relative_to(project_root)
    cfg.save_active_feature(project_root, str(rel))
    print(f"Active feature: {rel}")
    return 0


def cmd_trace(project_root: Path, ac_id: str, feature_ref: str | None) -> int:
    cfg = VibeConfig.load(project_root)
    feature_dir = resolve_feature(project_root, feature_ref, cfg)
    if not feature_dir:
        print("No feature found.")
        return 1

    ac_id = ac_id.upper()
    if not ac_id.startswith("AC-"):
        ac_id = f"AC-{ac_id.lstrip('0')}" if ac_id.isdigit() else ac_id

    print(f"## Trace {ac_id} — {feature_dir.name}\n")

    spec_path = feature_dir / "spec.md"
    if spec_path.exists():
        content = spec_path.read_text(encoding="utf-8", errors="replace")
        for line in content.splitlines():
            if ac_id in line.upper():
                print(f"Spec: {line.strip()}")

    tasks_path = feature_dir / "tasks.md"
    if tasks_path.exists():
        content = tasks_path.read_text(encoding="utf-8", errors="replace")
        print("\n### Tasks mapping")
        current_task = None
        for line in content.splitlines():
            if "**T-" in line:
                current_task = line.strip()
            if current_task and ac_id in line.upper():
                print(f"- {current_task}")
                if "Maps to:" in line or "Files:" in line:
                    print(f"  {line.strip()}")

    review_path = feature_dir / "review.md"
    if review_path.exists():
        content = review_path.read_text(encoding="utf-8", errors="replace")
        print("\n### Review evidence")
        for line in content.splitlines():
            if ac_id in line.upper():
                print(f"- {line.strip()}")

    return 0


def print_feature_report(r: dict[str, Any]) -> None:
    print(f"\n## {r['slug']}")
    print(f"Path: {r['path']}")
    print()
    print("| Phase | Gate | Status | File |")
    print("|-------|------|--------|------|")
    for p in r["phases"]:
        icon = {"pass": "✅", "fail": "❌", "in_progress": "🔄", "pending": "⬜"}.get(
            p["status"], "⬜"
        )
        print(f"| {p['phase']} | {p['gate']} | {icon} {p['status']} | {p.get('file', '')} |")
    print()
    print(f"**Current phase:** {r['current_phase']}")
    print(f"**Next command:** `{r['next_command']}`")
    if r["blockers"]:
        print("**Blockers:**")
        for b in r["blockers"]:
            print(f"  - {b}")
    if r.get("warnings"):
        print("**Warnings:**")
        for w in r["warnings"]:
            print(f"  - {w}")
    if not r["blockers"] and r.get("ship_ready"):
        print("**Status:** 🚀 Ship ready")


def cmd_validate(
    project_root: Path, feature: str | None, strict: bool
) -> int:
    cfg = VibeConfig.load(project_root)
    features = list_features(project_root, cfg)
    if not features:
        print("No features found.")
        print("Start with: vibe new <slug>  or  /vibe-idea <your idea>")
        return 1 if strict else 0

    if feature:
        matches = [f for f in features if feature in f.name]
        if not matches:
            print(f"Feature not found: {feature}")
            return 1
        features = matches

    exit_code = 0
    for fdir in features:
        r = validate_feature(fdir, cfg)
        print_feature_report(r)
        if r["blockers"]:
            exit_code = 1

    if strict and exit_code:
        return 1
    return exit_code


def main() -> int:
    parser = argparse.ArgumentParser(description="Vibe DevKit CLI")
    parser.add_argument("--project", "-p", default=".", help="Project root")
    sub = parser.add_subparsers(dest="command")

    v = sub.add_parser("validate", help="Validate pipeline artifacts and gates")
    v.add_argument("feature", nargs="?", help="Feature slug filter")
    v.add_argument("--strict", action="store_true", help="Exit 1 if any blocker")

    sub.add_parser("list-presets", help="List available presets")
    sub.add_parser("list-features", help="List feature folders")

    n = sub.add_parser("new", help="Create new feature folder from template")
    n.add_argument("slug", help="Feature slug (kebab-case)")
    n.add_argument("--title", "-t", help="Display title")

    sa = sub.add_parser("set-active", help="Set active feature in vibe.config.yaml")
    sa.add_argument("feature", help="Feature folder name or slug fragment")

    tr = sub.add_parser("trace", help="Trace AC to tasks and review")
    tr.add_argument("ac_id", help="AC id e.g. AC-001")
    tr.add_argument("--feature", "-f", help="Feature filter")

    gc = sub.add_parser("gate-check", help="Check prompt for gate (hook internal)")
    gc.add_argument("prompt", help="User prompt text")

    args = parser.parse_args()
    root = Path(args.project).resolve()

    if args.command == "validate":
        return cmd_validate(root, args.feature, args.strict)
    if args.command == "list-features":
        cfg = VibeConfig.load(root)
        for f in list_features(root, cfg):
            marker = ""
            if cfg.active_feature and f.name in cfg.active_feature:
                marker = " (active)"
            print(f"{f.name}{marker}")
        return 0
    if args.command == "list-presets":
        presets_dir = Path(__file__).resolve().parent.parent / "presets"
        if presets_dir.exists():
            for p in sorted(presets_dir.iterdir()):
                if p.is_dir() and p.name != "__pycache__":
                    print(p.name)
        return 0
    if args.command == "new":
        return cmd_new(root, args.slug, args.title)
    if args.command == "set-active":
        return cmd_set_active(root, args.feature)
    if args.command == "trace":
        return cmd_trace(root, args.ac_id, args.feature)
    if args.command == "gate-check":
        print(json.dumps(cmd_gate_check(root, args.prompt)))
        return 0

    parser.print_help()
    return 0


if __name__ == "__main__":
    sys.exit(main())
