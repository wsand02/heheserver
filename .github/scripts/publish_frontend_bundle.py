"""Publish generated data without checking out or executing PR code."""

import base64
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import tarfile
import time
from urllib.parse import quote
from urllib.request import Request, urlopen

DIST = "internal/frontend/dist/"
MANIFESTS = {"web/package.json", "web/pnpm-lock.yaml"}
MAX_BYTES = 32 * 1024 * 1024


class GitHub:
    def __init__(self, repository, token):
        self.root = os.environ.get("GITHUB_API_URL", "https://api.github.com") + "/repos/" + repository
        self.token = token

    def request(self, method, path, body=None):
        data = None if body is None else json.dumps(body).encode()
        request = Request(self.root + path, data=data, method=method, headers={
            "Authorization": "Bearer " + self.token,
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json",
            "X-GitHub-Api-Version": "2022-11-28",
        })
        with urlopen(request, timeout=30) as response:
            raw = response.read()
            return json.loads(raw) if raw else None


def dependency_only(files):
    def allowed(path):
        return path in MANIFESTS or path.startswith(DIST)
    return bool(files) and any(file["filename"] in MANIFESTS for file in files) and all(
        allowed(file["filename"]) and allowed(file.get("previous_filename", file["filename"]))
        for file in files
    )


def read_regular(path, limit):
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_size > limit:
        raise ValueError("Invalid bundle artifact file")
    return path.read_bytes()


def read_bundle(directory):
    directory = Path(directory)
    metadata = json.loads(read_regular(directory / "metadata.json", 4096))
    if (type(metadata.get("number")) is not int or metadata["number"] < 1
            or not re.fullmatch(r"[0-9a-f]{40}", metadata.get("sha", ""))):
        raise ValueError("Invalid PR metadata")
    archive = directory / "bundle.tar.gz"
    # No extraction: paths and contents remain data used only by the Git API.
    read_regular(archive, MAX_BYTES)
    files = {}
    total = 0
    with tarfile.open(archive, "r:gz") as bundle:
        for index, member in enumerate(bundle):
            if index >= 2048:
                raise ValueError("Too many generated files")
            name = member.name.removeprefix("./")
            if member.isdir():
                name = name.rstrip("/")
            parts = name.split("/")
            if member.isdir() and name in ("", "."):
                continue
            if (not name or "\\" in name or "\x00" in name or PurePosixPath(name).is_absolute()
                    or any(part in ("", ".", "..") or part.startswith(".") for part in parts)):
                raise ValueError("Invalid generated path")
            if member.isdir():
                continue
            total += member.size
            if not member.isfile() or total > MAX_BYTES or name in files:
                raise ValueError("Invalid generated file")
            files[name] = bundle.extractfile(member).read()
    if not files.get("index.html"):
        raise ValueError("Generated index.html is missing")
    return metadata, files


def blob_sha(content):
    return hashlib.sha1(b"blob " + str(len(content)).encode() + b"\0" + content).hexdigest()


def publish(api, repository, default_branch, run_id, metadata, files, ci_blob):
    run = api.request("GET", f"/actions/runs/{run_id}")
    if (run["path"].split("@")[0] != ".github/workflows/frontend-bundle.yml"
            or run["conclusion"] != "success"
            or run["event"] not in ("pull_request", "workflow_dispatch")
            or run["head_repository"]["full_name"] != repository):
        raise ValueError("Unexpected source workflow")
    if run["event"] == "pull_request" and run["head_sha"] != metadata["sha"]:
        raise ValueError("Artifact does not belong to this PR run")
    number, sha = metadata["number"], metadata["sha"]
    pr_path = f"/pulls/{number}"
    pr = api.request("GET", pr_path)
    if (pr["state"] != "open" or pr["head"]["sha"] != sha
            or (pr["head"].get("repo") or {}).get("full_name") != repository
            or pr["base"]["ref"] != default_branch
            or pr["head"]["ref"] == default_branch):
        print("Skipping closed, forked, or superseded PR.")
        return "skipped"
    changed = []
    for page in range(1, 32):
        batch = api.request("GET", f"{pr_path}/files?per_page=100&page={page}")
        changed.extend(batch)
        if len(batch) < 100:
            break
    else:
        raise ValueError("PR is too large for automatic bundle updates")
    if len(changed) != pr.get("changed_files", len(changed)):
        raise ValueError("PR file listing is incomplete")
    if not dependency_only(changed):
        print("Skipping PR with changes outside dependencies and generated assets.")
        return "skipped"
    parent = api.request("GET", f"/git/commits/{sha}")
    tree = api.request("GET", f"/git/trees/{parent['tree']['sha']}?recursive=1")
    if tree.get("truncated"):
        raise ValueError("Cannot validate a truncated repository tree")
    entries = {entry["path"]: entry for entry in tree["tree"]}
    if entries.get(".github/workflows/go.yml", {}).get("sha") != ci_blob:
        # Older dependency branches need the dispatchable CI workflow from master.
        # Merge the base via GitHub, then rebuild that new head rather than reuse
        # an artifact made from different sources.
        api.request("PUT", pr_path + "/update-branch", {"expected_head_sha": sha})
        for _ in range(30):
            updated = api.request("GET", pr_path)
            if updated["state"] != "open":
                return "skipped"
            if updated["head"]["sha"] != sha:
                api.request("POST", "/actions/workflows/frontend-bundle.yml/dispatches", {
                    "ref": default_branch, "inputs": {"pull_request": str(number)},
                })
                print("Updated dependency branch from base; rebuilding its new head.")
                return "refreshed"
            time.sleep(1)
        raise RuntimeError("Branch update has not completed; rerun this workflow")
    existing = {path[len(DIST):]: entry["sha"] for path, entry in entries.items()
                if path.startswith(DIST) and entry["type"] == "blob"}
    desired = {path: blob_sha(content) for path, content in files.items()}
    if existing == desired:
        print("Bundle already matches; no commit needed.")
        return "unchanged"
    if api.request("GET", pr_path)["head"]["sha"] != sha:
        print("PR changed during validation; skipping stale bundle.")
        return "skipped"
    generated = []
    for path, content in sorted(files.items()):
        blob = api.request("POST", "/git/blobs", {
            "content": base64.b64encode(content).decode(), "encoding": "base64",
        })
        generated.append({"path": path, "mode": "100644", "type": "blob", "sha": blob["sha"]})
    subtree = api.request("POST", "/git/trees", {"tree": generated})
    root = api.request("POST", "/git/trees", {"base_tree": parent["tree"]["sha"], "tree": [{
        "path": DIST.rstrip("/"), "mode": "040000", "type": "tree", "sha": subtree["sha"],
    }]})
    commit = api.request("POST", "/git/commits", {
        "message": "chore: rebuild embedded frontend after dependency update",
        "tree": root["sha"], "parents": [sha],
    })
    branch = pr["head"]["ref"]
    # Non-forced update fails if a concurrent push advanced the branch.
    api.request("PATCH", "/git/refs/heads/" + quote(branch, safe=""), {
        "sha": commit["sha"], "force": False,
    })
    api.request("POST", "/actions/workflows/go.yml/dispatches", {"ref": branch})
    print(f"Committed generated bundle to PR #{number} and dispatched CI.")
    return "published"


def main():
    event = json.loads(Path(os.environ["GITHUB_EVENT_PATH"]).read_text())
    repository = os.environ["GITHUB_REPOSITORY"]
    if event["repository"]["full_name"] != repository:
        raise ValueError("Unexpected repository")
    metadata, files = read_bundle(os.environ["BUNDLE_DIRECTORY"])
    publish(GitHub(repository, os.environ["GH_TOKEN"]), repository,
            event["repository"]["default_branch"], event["workflow_run"]["id"],
            metadata, files, blob_sha(Path(".github/workflows/go.yml").read_bytes()))


if __name__ == "__main__":
    main()
