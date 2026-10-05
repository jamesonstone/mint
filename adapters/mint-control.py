#!/usr/bin/env python3
"""Trusted event adapter. Mint owns all eligibility and state decisions."""
import json
import os
from pathlib import Path
import re
import subprocess
import sys


def gh(path):
    return json.loads(subprocess.check_output(["gh", "api", path]))


def mint(command, *arguments):
    return subprocess.run(["mint", "release", "production", command, *arguments], check=True)


def outputs(values):
    values = {key: str(value) for key, value in values.items()}
    if any("\n" in value or "\r" in value for value in values.values()):
        raise ValueError("multiline event identity")
    with open(os.environ["GITHUB_OUTPUT"], "a") as target:
        target.write("".join(f"{key}={value}\n" for key, value in values.items()))


def exact_number(value):
    if not re.fullmatch(r"[1-9][0-9]*", str(value)):
        raise ValueError("exact positive issue/PR number required")
    return int(value)


def source():
    pr_number = os.environ.get("HOTFIX_PR", "")
    if not pr_number:
        outputs({"source_sha": os.environ["GITHUB_SHA"], "kind": "normal", "source_pr": 0, "baseline_id": ""})
        return
    pr = gh(f"repos/{os.environ['GITHUB_REPOSITORY']}/pulls/{exact_number(pr_number)}")
    if not pr.get("merged") or pr["base"]["repo"]["full_name"] != os.environ["GITHUB_REPOSITORY"] or not pr["base"]["ref"].startswith("mint-hotfix-base/GH-"):
        raise ValueError("hotfix must be the exact merged production-base source PR")
    subprocess.run(["git", "fetch", "origin", pr["merge_commit_sha"]], check=True)
    result = Path(os.environ["RUNNER_TEMP"]) / "hotfix-version.json"
    mint("version-hotfix", "--pr", str(pr["number"]), "--output", str(result))
    version = json.loads(result.read_text())
    outputs({"source_sha": version["source_sha"], "kind": "hotfix", "source_pr": pr["number"], "baseline_id": version["baseline_id"], "version": version["version"]})


def promotion_request(event):
    sha = os.environ["GITHUB_SHA"]
    repo = os.environ["GITHUB_REPOSITORY"]
    pulls = []
    page = 1
    while True:
        batch = gh(f"repos/{repo}/commits/{sha}/pulls?per_page=100&page={page}")
        pulls.extend(batch)
        if len(batch) < 100:
            break
        page += 1
    matches = [pr for pr in pulls if pr.get("merged_at") and pr["merge_commit_sha"] == sha and pr["base"]["repo"]["full_name"] == repo and "<!-- mint:proposal:" in (pr.get("body") or "") and ":status -->" not in pr["body"]]
    if not matches:
        outputs({"promote": "false"})
        return
    if len(matches) != 1:
        raise ValueError("ambiguous merged release proposal")
    result = Path(os.environ["RUNNER_TEMP"]) / "intent.json"
    mint("intent", "--pr", str(matches[0]["number"]), "--merge-sha", sha, "--output", str(result))
    intent = json.loads(result.read_text())
    mint("start", "--intent-id", intent["id"], "--run-id", os.environ["GITHUB_RUN_ID"], "--output", str(result))
    outputs({"promote": "true"})
    subprocess.run([sys.executable, str(Path(__file__).with_name("mint-manifest.py")), "intent-outputs", str(result)], check=True)


def request_hotfix(issue_number):
    repo = os.environ["GITHUB_REPOSITORY"]
    issue = gh(f"repos/{repo}/issues/{exact_number(issue_number)}")
    body = issue.get("body") or ""
    def field(heading):
        match = re.search(r"^### " + re.escape(heading) + r"\s*\n(.*?)(?=^### |\Z)", body, re.M | re.S)
        if not match:
            raise ValueError(f"missing {heading}")
        return match.group(1).strip()
    baseline = field("Verified production baseline ID")
    fixes = field("Reviewed fix SHAs")
    fixes = [] if fixes == "AUTHORED" else fixes.split()
    if any(not re.fullmatch(r"[a-f0-9]{40}", fix) for fix in fixes):
        raise ValueError("fixes must be full lowercase single-parent SHAs")
    request = Path(os.environ["RUNNER_TEMP"]) / "hotfix-request.json"
    request.write_text(json.dumps({"Issue": issue["number"], "BaselineID": baseline, "Fixes": fixes}))
    mint("prepare-hotfix", "--input", str(request))


def reconcile(event, config):
    operation = os.environ.get("CONTROL_OPERATION", "reconcile")
    if operation == "hotfix":
        request_hotfix(os.environ["REQUEST_ISSUE"])
        return
    if operation == "rollback":
        mint("propose-rollback", "--pin", os.environ["ROLLBACK_SOURCE"], "--summary", os.environ["RELEASE_SUMMARY"])
        return
    if operation == "publish":
        mint("publish", "--intent-id", os.environ["REQUEST_INTENT"])
        mint("report", "--intent-id", os.environ["REQUEST_INTENT"])
        return
    if os.environ["GITHUB_EVENT_NAME"] in ["pull_request", "pull_request_target"]:
        pr = gh(f"repos/{config['Repository']}/pulls/{exact_number(event['pull_request']['number'])}")
        if pr["head"]["repo"]["full_name"] != config["Repository"] or pr["base"]["repo"]["full_name"] != config["Repository"]:
            raise ValueError("foreign proposal")
        # A reviewed production-base merge does not push main. Build it through
        # the trusted default-branch producer explicitly, using only its PR id.
        if pr.get("merged") and "<!-- mint:hotfix-source:" in (pr.get("body") or ""):
            number = exact_number(pr["number"])
            workflow = config["BuildWorkflow"].split("/")[-1]
            subprocess.run(["gh", "workflow", "run", workflow, "--ref", config["DefaultBranch"], "-f", f"hotfix_pr={number}"], check=True)
            return
        if "<!-- mint:proposal:" not in (pr.get("body") or "") or ":status -->" in pr["body"] or pr.get("merged"):
            return
        if event["action"] == "closed":
            action = "close"
        elif event["action"] == "reopened":
            action = "reopen"
        else:
            action = "candidate"
        state = gh(f"repos/{os.environ['GITHUB_REPOSITORY']}/pulls/{pr['number']}")
        if state["head"]["repo"]["full_name"] != os.environ["GITHUB_REPOSITORY"]:
            raise ValueError("foreign proposal")
        # Match durable PR identity; do not parse title or trust event labels.
        snapshot = subprocess.check_output(["mint", "release", "production", "status"])
        journal = json.loads(snapshot)
        kind = next((k for k, p in journal["proposals"].items() if p["pr"] == pr["number"]), None)
        if kind is None:
            raise ValueError("unregistered release proposal")
        mint("propose", "--kind", kind, "--event", action)
        return
    run = event.get("workflow_run")
    if run:
        run = gh(f"repos/{config['Repository']}/actions/runs/{exact_number(run['id'])}")
        if run["repository"]["full_name"] != config["Repository"] or run["head_repository"]["full_name"] != config["Repository"] or run["status"] != "completed":
            raise ValueError("foreign or incomplete callback")
    if run and run["path"].split("@")[0] == config["PromotionWorkflow"]:
        # Exact merge SHA resolves the frozen journal intent; no latest-source lookup.
        snapshot = json.loads(subprocess.check_output(["mint", "release", "production", "status"]))
        matches = [i for i in snapshot["intents"].values() if i["merge_sha"] == run["head_sha"] and i.get("deployment_run_id") == run["id"]]
        if not matches:
            return
        if len(matches) != 1:
            raise ValueError("ambiguous completed deployment intent")
        intent = matches[0]
        mint("finish", "--intent-id", intent["id"], "--run-id", str(run["id"]))
        if run["conclusion"] == "success" and intent.get("kind") != "rollback":
            mint("publish", "--intent-id", intent["id"])
        mint("report", "--intent-id", intent["id"])
        if run["conclusion"]=="success" and intent.get("kind")=="hotfix":
            mint("propose","--kind","normal")
        return
    if run and (run["path"].split("@")[0] != config["BuildWorkflow"] or run["conclusion"] != "success"):
        return
    mint("scan")
    snapshot=json.loads(subprocess.check_output(["mint","release","production","status"]))
    if run:
        # Direct registration surfaces missing/invalid evidence for this specific event.
        artifacts=gh(f"repos/{os.environ['GITHUB_REPOSITORY']}/actions/runs/{run['id']}/artifacts?per_page=100")
        if not any(a["name"]=="mint-candidate" for a in artifacts["artifacts"]):
            paths=subprocess.check_output(["git","diff-tree","--root","-m","--first-parent","--no-commit-id","--name-only","-r",run["head_sha"]]).decode().splitlines()
            if paths and all(p in set(config["ControlPaths"]) for p in paths):
                return
            raise ValueError("successful source producer did not attest its artifact")
        output=Path(os.environ["RUNNER_TEMP"])/"candidate.json"
        mint("candidate", "--run-id", str(run["id"]),"--output",str(output))
        candidate=json.loads(output.read_text())
        if candidate["kind"]=="hotfix":
            mint("propose","--kind","hotfix","--pin",candidate["source_sha"])
            return
    if snapshot.get("in_flight"):
        return
    prior=snapshot["proposals"].get("normal",{})
    if prior.get("state") in {"merged","deployment_failed","publication_pending"}:
        return
    mint("propose", "--kind", "normal")


if __name__ == "__main__":
    operation = sys.argv[1]
    role = {"source": "source", "promote": "promote", "reconcile": "control"}.get(operation)
    if role is None:
        raise ValueError("unsupported event operation")
    config = json.loads(subprocess.check_output(["mint", "release", "production", "adapter-policy", "--role", role, "--event", os.environ["GITHUB_EVENT_NAME"], "--run-id", os.environ["GITHUB_RUN_ID"]]))
    if config["Repository"] != os.environ["GITHUB_REPOSITORY"]:
        raise ValueError("repository does not match authenticated policy")
    event = json.loads(Path(os.environ["GITHUB_EVENT_PATH"]).read_text())
    if sys.argv[1] == "source":
        source()
    elif sys.argv[1] == "promote":
        promotion_request(event)
    elif sys.argv[1] == "reconcile":
        reconcile(event, config)
    else:
        raise ValueError("unsupported event operation")
