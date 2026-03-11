# Mini DSL spec (NO coordinates) to generate grafana dashboard json configuration.
# Layout is computed by the compiler.
#
# Mental model:
# - You describe intent: sections + panels.
# - Compiler packs panels into Grafana’s 24-col grid (flow layout) and emits:
#   - Grafana dashboard JSON (classic model: panels + gridPos)
#   - Optional provisioning YAML (points Grafana to the JSON)

from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any, Sequence, Union

# ----------------------------
# DSL Core Types (spec only)
# ----------------------------

@dataclass
class Target:
    expr: str
    legend: str = ""

@dataclass
class Panel:
    kind: str                 # "timeseries" | "stat" | "row"
    title: str
    targets: List[Target] = field(default_factory=list)
    # sizing is the only "layout" the author should care about
    w: int = 8
    h: int = 7
    # optional: for stat thresholds etc. (raw grafana json fragments)
    field_defaults: Optional[Dict[str, Any]] = None
    # teaching aid: shown in docs/README, not necessarily in Grafana
    how_to_read: Optional[str] = None
    look_for: Optional[str] = None
    common_traps: Optional[str] = None

@dataclass
class Section:
    title: str
    panels: List[Panel]
    # optional “intro” text for teaching (not necessarily embedded in Grafana)
    teaching_note: Optional[str] = None

@dataclass
class Dashboard:
    title: str
    uid: str
    tags: List[str]
    refresh: str = "5s"
    time_from: str = "now-30m"
    time_to: str = "now"
    schema_version: int = 39
    version: int = 1
    sections: List[Section] = field(default_factory=list)

# ----------------------------
# Helper constructors
# ----------------------------

def row(title: str, teaching_note: Optional[str] = None) -> Section:
    # In the compiled Grafana dashboard, each section becomes a "row" panel + its children.
    return Section(title=title, panels=[], teaching_note=teaching_note)

def ts(
        title: str,
        targets: Sequence[Target],
        w: int = 8,
        h: int = 7,
        how_to_read: Optional[str] = None,
        look_for: Optional[str] = None,
        common_traps: Optional[str] = None,
) -> Panel:
    return Panel(
        kind="timeseries",
        title=title,
        targets=list(targets),
        w=w,
        h=h,
        how_to_read=how_to_read,
        look_for=look_for,
        common_traps=common_traps,
    )

def stat(
        title: str,
        targets: Sequence[Target],
        w: int = 6,
        h: int = 6,
        thresholds_steps: Optional[List[Dict[str, Union[str, float, int]]]] = None,
        how_to_read: Optional[str] = None,
        look_for: Optional[str] = None,
        common_traps: Optional[str] = None,
) -> Panel:
    field_defaults = None
    if thresholds_steps is not None:
        field_defaults = {
            "thresholds": {"steps": thresholds_steps}
        }
    return Panel(
        kind="stat",
        title=title,
        targets=list(targets),
        w=w,
        h=h,
        field_defaults=field_defaults,
        how_to_read=how_to_read,
        look_for=look_for,
        common_traps=common_traps,
    )
