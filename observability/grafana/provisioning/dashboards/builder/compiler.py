import json
from typing import Dict, Any, List

from spec import DASH
from dsl import Panel, Section, Target

GRID_W = 24

def flow_layout(sections: List[Section]) -> List[Dict[str, Any]]:
    """
    Packs panels into Grafana 24-col grid, left->right, wrap->next line.
    Adds a 1-row 'row' panel before each section.
    Returns Grafana 'panels' JSON list with gridPos populated.
    """
    panels_json: List[Dict[str, Any]] = []
    next_id = 1
    y = 0

    for sec in sections:
        # Section header row
        panels_json.append({
            "id": next_id,
            "type": "row",
            "title": sec.title,
            "gridPos": {"h": 1, "w": 24, "x": 0, "y": y},
        })
        next_id += 1
        y += 1

        x = 0
        line_h = 0

        for p in sec.panels:
            # Wrap if it doesn't fit
            if x + p.w > GRID_W:
                y += line_h
                x = 0
                line_h = 0

            panel_obj: Dict[str, Any] = {
                "id": next_id,
                "type": p.kind,
                "title": p.title,
                "gridPos": {"h": p.h, "w": p.w, "x": x, "y": y},
                "targets": [
                    {"expr": t.expr, "legendFormat": (t.legend or p.title)}
                    for t in p.targets
                ],
            }

            # For stat panels: thresholds etc.
            if p.field_defaults is not None:
                panel_obj["fieldConfig"] = {"defaults": p.field_defaults}

            panels_json.append(panel_obj)
            next_id += 1

            x += p.w
            line_h = max(line_h, p.h)

        # Move down after section
        y += line_h

    return panels_json


def dashboard_to_grafana_json() -> Dict[str, Any]:
    return {
        "title": DASH.title,
        "uid": DASH.uid,
        "schemaVersion": DASH.schema_version,
        "version": DASH.version,
        "refresh": DASH.refresh,
        "time": {"from": DASH.time_from, "to": DASH.time_to},
        "tags": DASH.tags,
        "panels": flow_layout(DASH.sections),
    }


def main():
    out = dashboard_to_grafana_json()
    with open("../dashboard.json", "w") as f:
        json.dump(out, f, indent=2)
    print("Wrote dashboard.json")


if __name__ == "__main__":
    main()