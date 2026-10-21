"""Fake Google Analytics Data API and Search Console API, answering with sample data for the screenshots.

Usage: fakegoogle.py <port> [requests.log]. It gives a token to any keyfile, and logs the requests.
"""
import datetime as dt
import itertools
import json
import math
import random
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer

LOG = open(sys.argv[2], "a") if len(sys.argv) > 2 else sys.stderr

VALUES = {
    "pagePath": ["/", "/blog/", "/blog/terminal-dashboards/", "/blog/go-generics/", "/about/",
                 "/blog/vim-tips/", "/projects/", "/blog/tag/go/"],
    "landingPage": ["/", "/blog/terminal-dashboards/", "/blog/go-generics/"],
    "pageReferrer": ["https://www.google.com/", "https://news.ycombinator.com/", "https://github.com/"],
    "sessionSource": ["google", "(direct)", "duckduckgo", "github.com", "news.ycombinator.com", "bing", "reddit.com"],
    "newVsReturning": ["new", "returning"],
    "deviceCategory": ["desktop", "mobile", "tablet"],
    "country": ["United States", "Germany", "France", "India", "United Kingdom", "Canada"],
    "city": ["Berlin", "Paris", "New York", "London"],
}
WEIGHTS = {"desktop": 1.0, "mobile": 0.6, "tablet": 0.1, "new": 1.0, "returning": 0.45, "(direct)": 0.6,
           "google": 1.0, "duckduckgo": 0.25, "github.com": 0.2, "news.ycombinator.com": 0.15, "bing": 0.1,
           "reddit.com": 0.08, "/": 1.0, "/blog/": 0.7, "/blog/terminal-dashboards/": 0.55,
           "/blog/go-generics/": 0.4, "/about/": 0.2, "/blog/vim-tips/": 0.3, "/projects/": 0.15,
           "/blog/tag/go/": 0.05, "United States": 1.0, "Germany": 0.5, "France": 0.4, "India": 0.35,
           "United Kingdom": 0.3, "Canada": 0.2}
BASE = {"sessions": 420, "totalUsers": 310, "newUsers": 200, "activeUsers": 300, "screenPageViews": 900,
        "engagedSessions": 260, "sessionsPerUser": 1.35, "engagementRate": 0.62, "bounceRate": 0.38,
        "averageSessionDuration": 94.2}


def days(start, end):
    d = dt.date.fromisoformat(start)
    e = dt.date.fromisoformat(end)
    while d <= e:
        yield d
        d += dt.timedelta(days=1)


def time_values(name, start, end):
    ds = list(days(start, end))
    if name == "date":
        return [d.strftime("%Y%m%d") for d in ds]
    if name == "yearMonth":
        return sorted({d.strftime("%Y%m") for d in ds})
    if name == "year":
        return sorted({d.strftime("%Y") for d in ds})
    return None


def contains(expr):
    """Values a filter expression keeps, as a predicate."""
    if not expr:
        return lambda name, v: True
    if "filter" in expr:
        f = expr["filter"]
        return lambda name, v: name != f["fieldName"] or f["stringFilter"]["value"].lower() in v.lower()
    if "notExpression" in expr:
        p = contains(expr["notExpression"])
        return lambda name, v: not p(name, v) if name == field(expr["notExpression"]) else True
    if "orGroup" in expr:
        ps = [contains(e) for e in expr["orGroup"]["expressions"]]
        return lambda name, v: any(p(name, v) for p in ps)
    if "andGroup" in expr:
        ps = [contains(e) for e in expr["andGroup"]["expressions"]]
        return lambda name, v: all(p(name, v) for p in ps)
    return lambda name, v: True


def field(expr):
    if "filter" in expr:
        return expr["filter"]["fieldName"]
    for k in ("notExpression",):
        if k in expr:
            return field(expr[k])
    for k in ("orGroup", "andGroup"):
        if k in expr:
            return field(expr[k]["expressions"][0])
    return None


def metric_value(metric, key, scale):
    rnd = random.Random(json.dumps(key) + metric.get("name", ""))
    name = metric.get("name")
    if metric.get("expression") == "sessions-engagedSessions":
        v = (BASE["sessions"] - BASE["engagedSessions"]) * scale
    elif metric.get("expression") == "averageSessionDuration*sessions":
        v = BASE["averageSessionDuration"] * BASE["sessions"] * scale
    else:
        v = BASE.get(name, 100)
        if name not in ("sessionsPerUser", "engagementRate", "bounceRate", "averageSessionDuration"):
            v *= scale
    v *= rnd.uniform(0.8, 1.2)
    if isinstance(BASE.get(name), float):
        return f"{v:.4f}"
    return str(max(int(v), 0))


def run_report(body):
    (dr,) = body["dateRanges"]
    dims = [d["name"] for d in body.get("dimensions", [])]
    keep = contains(body.get("dimensionFilter"))
    n_days = len(list(days(dr["startDate"], dr["endDate"])))
    axes = []
    for name in dims:
        values = time_values(name, dr["startDate"], dr["endDate"]) or VALUES.get(name, ["(not set)"])
        axes.append([v for v in values if keep(name, v)])
    rows = []
    for combo in itertools.product(*axes) if dims else [()]:
        scale = 1.0
        for name, v in zip(dims, combo):
            if time_values(name, dr["startDate"], dr["endDate"]) is not None:
                # A growing trend over time, with less traffic on weekends.
                if name == "date":
                    d = dt.datetime.strptime(v, "%Y%m%d").date()
                    scale *= (0.65 if d.weekday() >= 5 else 1.0) * (1 + 0.02 * (d - dt.date(2026, 1, 1)).days / 30)
                elif name == "yearMonth":
                    scale *= 30 * (1 + 0.08 * (int(v[:4]) * 12 + int(v[4:]) - 2026 * 12))
                else:
                    scale *= 365
            else:
                scale *= WEIGHTS.get(v, 0.3)
        if not any(time_values(n, dr["startDate"], dr["endDate"]) is not None for n in dims):
            scale *= n_days
        rows.append({
            "dimensionValues": [{"value": v} for v in combo],
            "metricValues": [{"value": metric_value(m, combo, scale)} for m in body["metrics"]],
        })
    for o in reversed(body.get("orderBys", [])):
        if "metric" in o:
            names = [m["name"] for m in body["metrics"]]
            i = names.index(o["metric"]["metricName"])
            rows.sort(key=lambda r: float(r["metricValues"][i]["value"]), reverse=o.get("desc", False))
        elif "dimension" in o:
            i = dims.index(o["dimension"]["dimensionName"])
            rows.sort(key=lambda r: r["dimensionValues"][i]["value"], reverse=o.get("desc", False))
    if body.get("limit"):
        rows = rows[: int(body["limit"])]
    return {
        "dimensionHeaders": [{"name": d} for d in dims],
        "metricHeaders": [{"name": m["name"]} for m in body["metrics"]],
        "rows": rows,
        "rowCount": len(rows),
    }


GSC = {
    "query": ["terminal dashboard", "devdash", "go generics tutorial", "vim tips", "tcell go",
              "terminal dashboard linux", "go tui library", "google analytics terminal", "cli dashboard"],
    "page": ["https://example.com/blog/terminal-dashboards/", "https://example.com/blog/go-generics/",
             "https://example.com/", "https://example.com/blog/vim-tips/", "https://example.com/blog/tag/go/",
             "https://example.com/projects/", "https://example.com/about/"],
    "country": ["usa", "deu", "fra", "ind", "gbr"],
    "device": ["DESKTOP", "MOBILE", "TABLET"],
}


def search_analytics(body):
    (dim,) = body["dimensions"]
    filters = [f for g in body.get("dimensionFilterGroups", []) for f in g.get("filters") or []]
    rows = []
    for i, k in enumerate(GSC.get(dim, ["(other)"])):
        ok = True
        for f in filters:
            if f["dimension"] != dim:
                continue
            hit = f["expression"].lower() in k.lower()
            ok &= hit if f.get("operator", "contains") == "contains" else not hit
        if not ok:
            continue
        rnd = random.Random(k)
        clicks = int(900 * math.exp(-0.45 * i) * rnd.uniform(0.85, 1.15))
        impressions = int(clicks * rnd.uniform(9, 30))
        rows.append({"keys": [k], "clicks": clicks, "impressions": impressions,
                     "ctr": clicks / impressions, "position": rnd.uniform(1.5, 12)})
    rows.sort(key=lambda r: -r["clicks"])
    return {"rows": rows[: int(body.get("rowLimit") or 1000)]}


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        body = self.rfile.read(int(self.headers.get("Content-Length", 0)))
        LOG.write(f"{self.path} {body.decode(errors='replace')}\n")
        LOG.flush()
        path = self.path.split("?")[0]
        if path.startswith("/token"):
            out = {"access_token": "fake", "token_type": "Bearer", "expires_in": 3600}
        else:
            req = json.loads(body or b"{}")
            if path.endswith(":runRealtimeReport"):
                out = {"rows": [{"metricValues": [{"value": "17"}]}]}
            elif path.endswith(":runReport"):
                out = run_report(req)
            elif "searchAnalytics/query" in path:
                out = search_analytics(req)
            else:
                self.send_error(404)
                return
        data = json.dumps(out).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def log_message(self, *args):
        pass


HTTPServer(("127.0.0.1", int(sys.argv[1])), Handler).serve_forever()
