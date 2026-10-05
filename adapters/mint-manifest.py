#!/usr/bin/env python3
"""Typed application manifests; no release selection or lifecycle policy lives here."""
import hashlib
import json
import os
from pathlib import Path
import re
import sys


def required(name):
    value = os.environ.get(name, "")
    if not value:
        raise ValueError(f"{name} is required")
    return value


def digest(value):
    return "sha256:" + hashlib.sha256(value).hexdigest()


def configuration():
    # Freeze standing repository configuration and adapter/template identity.
    variables = json.loads(required("RELEASE_VARIABLES"))
    excluded = json.loads(os.environ.get("RELEASE_CONFIG_EXCLUDE_VARIABLES", "[]"))
    files = json.loads(required("RELEASE_CONFIG_FILES"))
    if not isinstance(variables, dict) or not isinstance(excluded, list) or not isinstance(files, list):
        raise ValueError("configuration requires variables object and file/exclusion lists")
    if any(not isinstance(key, str) for key in excluded + files) or len(set(files)) != len(files):
        raise ValueError("configuration paths and exclusions must be unique strings")
    variables = {key: value for key, value in variables.items() if not key.startswith("MINT_") and key not in excluded}
    identities = {}
    for name in files:
        path = Path(name)
        if path.is_absolute() or ".." in path.parts or not path.parts:
            raise ValueError("configuration must use repository-relative files")
        if any(parent.is_symlink() for parent in [path, *path.parents]) or not path.is_file():
            raise ValueError("configuration must contain regular files, not links")
        identities[name] = digest(path.read_bytes())
    identity = {"variables": variables, "files": identities}
    return digest(json.dumps(identity, sort_keys=True, separators=(",", ":")).encode())


def exact(value, pattern, label):
    if not re.fullmatch(pattern, value):
        raise ValueError(f"invalid {label}")
    return value


def candidate():
    source = exact(required("SOURCE_SHA"), r"[a-f0-9]{40}", "source SHA")
    version = exact(required("VERSION_TAG"), r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)", "version")
    artifact_digest = exact(required("ARTIFACT_DIGEST"), r"sha256:[a-f0-9]{64}", "artifact digest")
    return {"repository": required("GITHUB_REPOSITORY"), "environment": required("RELEASE_ENVIRONMENT"), "source_sha": source,
            "version": version, "kind": os.environ.get("CANDIDATE_KIND", "normal"), "source_date": "",
            "artifact": {"reference": required("ARTIFACT_REFERENCE"), "digest": artifact_digest,
                         "configuration_sha256": configuration()}, "run_id": int(os.environ.get("SOURCE_BUILD_RUN_ID") or required("GITHUB_RUN_ID")),
            "run_url": os.environ.get("SOURCE_BUILD_RUN_URL", ""), "changes": [], "baseline_id": os.environ.get("BASELINE_ID", ""),
            "source_pr": int(os.environ.get("SOURCE_PR", "0")), "control_only": False}


def main():
    operation = sys.argv[1]
    if operation == "candidate":
        output = candidate()
    elif operation == "baseline":
        selected = json.loads(Path(sys.argv[2]).read_text())
        if not selected["run_url"]:
            raise ValueError("verified baseline requires exact source build evidence")
        output = {"id": "baseline-" + selected["source_sha"] + "-" + selected["artifact"]["digest"][7:],
                  "candidate": selected, "deployment_url": required("GITHUB_SERVER_URL") + "/" + required("GITHUB_REPOSITORY") + "/actions/runs/" + required("GITHUB_RUN_ID"),
                  "shipped": [], "publication_pending": False}
        Path(sys.argv[3]).write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")
        return
    elif operation == "configuration":
        print(configuration())
        return
    elif operation == "verify-configuration":
        if configuration() != required("EXPECTED_CONFIGURATION"):
            raise ValueError("configuration changed since this artifact was built; review a fresh source build")
        return
    elif operation == "deployment":
        if required("DEPLOYMENT_VERIFIED") != "true":
            raise ValueError("provider must verify the runtime before attesting deployment")
        output = {"intent_id": required("INTENT_ID"), "source_sha": required("SOURCE_SHA"),
                  "artifact_digest": required("ARTIFACT_DIGEST"), "configuration_sha256": configuration(), "verified": True}
        if output["configuration_sha256"] != required("EXPECTED_CONFIGURATION"):
            raise ValueError("verified deployment configuration differs from reviewed intent")
    elif operation == "intent-outputs":
        intent = json.loads(Path(sys.argv[2]).read_text())
        selected = intent["candidate"]
        values = {"intent_id": intent["id"], "source_sha": selected["source_sha"], "version": selected["version"],
                  "reference": selected["artifact"]["reference"], "digest": selected["artifact"]["digest"],
                  "configuration": selected["artifact"]["configuration_sha256"], "status": intent["status"], "kind": intent.get("kind", selected["kind"])}
        if any("\n" in value or "\r" in value for value in values.values()):
            raise ValueError("multiline intent identity")
        with open(required("GITHUB_OUTPUT"), "a") as target:
            target.write("".join(f"{key}={value}\n" for key, value in values.items()))
        return
    else:
        raise ValueError("unsupported manifest operation")
    Path(sys.argv[2]).write_text(json.dumps(output, sort_keys=True, indent=2) + "\n")


if __name__ == "__main__":
    main()
