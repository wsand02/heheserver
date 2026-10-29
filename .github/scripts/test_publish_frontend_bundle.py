import base64
import copy
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest

from publish_frontend_bundle import DIST, blob_sha, dependency_only, publish, read_bundle

SHA = "1" * 40
METADATA = {"number": 7, "sha": SHA}
FILES = {"index.html": b'<script src="/assets/new.js"></script>', "assets/new.js": b"compiled"}
REPOSITORY = "owner/repo"


class FakeGitHub:
    def __init__(self):
        self.calls = []
        self.pr = {"state": "open", "head": {
            "sha": SHA, "ref": "dependabot/npm_and_yarn/web/react", "repo": {"full_name": REPOSITORY},
        }, "base": {"ref": "master"}}
        self.run = {"path": ".github/workflows/frontend-bundle.yml", "conclusion": "success",
                    "event": "pull_request", "head_sha": SHA, "head_repository": {"full_name": REPOSITORY}}
        self.changed = [{"filename": "web/pnpm-lock.yaml"}]
        self.entries = [
            {"path": ".github/workflows/go.yml", "sha": "ci", "type": "blob"},
            {"path": DIST + "assets/old.js", "sha": "old", "type": "blob"},
        ]
        self.truncated = False
        self.stale_on_recheck = False
        self.reads = 0

    def request(self, method, path, body=None):
        self.calls.append((method, path, body))
        if path == "/actions/runs/42":
            return self.run
        if method == "GET" and path == "/pulls/7":
            self.reads += 1
            if self.stale_on_recheck and self.reads > 1:
                self.pr["head"]["sha"] = "2" * 40
            return copy.deepcopy(self.pr)
        if path.startswith("/pulls/7/files?"):
            return self.changed
        if path == f"/git/commits/{SHA}":
            return {"tree": {"sha": "parent-tree"}}
        if path == "/git/trees/parent-tree?recursive=1":
            return {"tree": self.entries, "truncated": self.truncated}
        if method == "PUT" and path == "/pulls/7/update-branch":
            self.pr["head"]["sha"] = "3" * 40
            return {}
        if method == "POST" and path == "/git/blobs":
            return {"sha": blob_sha(base64.b64decode(body["content"]))}
        if method == "POST" and path == "/git/trees":
            return {"sha": "root-tree" if "base_tree" in body else "bundle-tree"}
        if method == "POST" and path == "/git/commits":
            return {"sha": "4" * 40}
        if method == "PATCH" or path.endswith("/dispatches"):
            return None
        raise AssertionError(f"Unexpected API call: {method} {path}")


class PublisherTests(unittest.TestCase):
    def publish(self, api):
        return publish(api, REPOSITORY, "master", 42, METADATA, FILES, "ci")

    def test_publishes_only_asset_subtree_and_dispatches_ci(self):
        api = FakeGitHub()
        self.assertEqual(self.publish(api), "published")
        trees = [body for method, path, body in api.calls if method == "POST" and path == "/git/trees"]
        self.assertEqual({entry["path"] for entry in trees[0]["tree"]}, set(FILES))
        self.assertNotIn("base_tree", trees[0])  # Replaces stale hashed files.
        self.assertEqual(trees[1], {"base_tree": "parent-tree", "tree": [{
            "path": DIST.rstrip("/"), "mode": "040000", "type": "tree", "sha": "bundle-tree",
        }]})
        commit = next(body for method, path, body in api.calls if path == "/git/commits")
        self.assertEqual(commit["parents"], [SHA])
        patch = next(call for call in api.calls if call[0] == "PATCH")
        self.assertTrue(patch[1].endswith("dependabot%2Fnpm_and_yarn%2Fweb%2Freact"))
        self.assertEqual(patch[2], {"sha": "4" * 40, "force": False})
        self.assertEqual(api.calls[-1], ("POST", "/actions/workflows/go.yml/dispatches", {"ref": api.pr["head"]["ref"]}))

    def test_unchanged_bundle_needs_no_commit(self):
        api = FakeGitHub()
        api.entries = [entry for entry in api.entries if not entry["path"].startswith(DIST)] + [
            {"path": DIST + path, "sha": blob_sha(data), "type": "blob"} for path, data in FILES.items()
        ]
        self.assertEqual(self.publish(api), "unchanged")
        self.assertTrue(all(method == "GET" for method, _, _ in api.calls))

    def test_skips_forks_closed_prs_and_new_heads(self):
        for kind in ("fork", "closed", "new-head", "base-branch"):
            with self.subTest(kind=kind):
                api = FakeGitHub()
                if kind == "fork":
                    api.pr["head"]["repo"]["full_name"] = "attacker/fork"
                elif kind == "closed":
                    api.pr["state"] = "closed"
                elif kind == "new-head":
                    api.pr["head"]["sha"] = "2" * 40
                else:
                    api.pr["head"]["ref"] = "master"
                self.assertEqual(self.publish(api), "skipped")
                self.assertTrue(all(method == "GET" for method, _, _ in api.calls))

    def test_skips_stale_head_during_validation(self):
        api = FakeGitHub()
        api.stale_on_recheck = True
        self.assertEqual(self.publish(api), "skipped")
        self.assertTrue(all(method == "GET" for method, _, _ in api.calls))

    def test_rejects_artifact_from_wrong_workflow_or_commit(self):
        for field, value in (("path", "other.yml"), ("head_sha", "2" * 40), ("event", "push")):
            with self.subTest(field=field):
                api = FakeGitHub()
                api.run[field] = value
                with self.assertRaises(ValueError):
                    self.publish(api)

    def test_manual_dispatch_is_supported(self):
        api = FakeGitHub()
        api.run.update(event="workflow_dispatch", head_sha="default-branch-sha")
        self.assertEqual(self.publish(api), "published")

    def test_refreshes_old_branch_before_rebuilding(self):
        api = FakeGitHub()
        api.entries[0]["sha"] = "old-ci-without-workflow-dispatch"
        self.assertEqual(self.publish(api), "refreshed")
        self.assertIn(("PUT", "/pulls/7/update-branch", {"expected_head_sha": SHA}), api.calls)
        self.assertEqual(api.calls[-1], ("POST", "/actions/workflows/frontend-bundle.yml/dispatches", {
            "ref": "master", "inputs": {"pull_request": "7"},
        }))
        self.assertFalse(any(path == "/git/blobs" for _, path, _ in api.calls))

    def test_rejects_source_changes_and_renames(self):
        for file in ({"filename": "web/src/App.tsx"}, {"filename": DIST + "code.js", "previous_filename": "main.go"}):
            api = FakeGitHub()
            api.changed.append(file)
            self.assertEqual(self.publish(api), "skipped")
        self.assertFalse(dependency_only([{"filename": DIST + "index.html"}]))

    def test_rejects_truncated_tree(self):
        api = FakeGitHub()
        api.truncated = True
        with self.assertRaises(ValueError):
            self.publish(api)

    def test_rejects_incomplete_pr_file_list(self):
        api = FakeGitHub()
        api.pr["changed_files"] = 3001
        with self.assertRaises(ValueError):
            self.publish(api)


class ArtifactTests(unittest.TestCase):
    def archive(self, entries):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        directory = Path(temporary.name)
        (directory / "metadata.json").write_text(json.dumps(METADATA))
        with tarfile.open(directory / "bundle.tar.gz", "w:gz") as archive:
            for name, content, kind in entries:
                entry = tarfile.TarInfo(name)
                entry.type = kind
                entry.size = len(content) if kind == tarfile.REGTYPE else 0
                entry.linkname = "../../outside"
                archive.addfile(entry, io.BytesIO(content))
        return directory

    def test_reads_bundle_as_data_without_extracting(self):
        directory = self.archive([
            ("./", b"", tarfile.DIRTYPE), ("./assets/", b"", tarfile.DIRTYPE),
            *[(f"./{path}", data, tarfile.REGTYPE) for path, data in FILES.items()],
        ])
        self.assertEqual(read_bundle(directory), (METADATA, FILES))
        self.assertFalse((directory / "assets").exists())

    def test_committed_bundle_roundtrip(self):
        root = Path(__file__).resolve().parents[2] / "internal/frontend/dist"
        directory = self.archive([])
        with tarfile.open(directory / "bundle.tar.gz", "w:gz") as archive:
            archive.add(root, arcname=".")
        _, generated = read_bundle(directory)
        expected = {path.relative_to(root).as_posix(): path.read_bytes()
                    for path in root.rglob("*") if path.is_file()}
        self.assertEqual(generated, expected)

    def test_rejects_paths_links_missing_index_and_duplicate_files(self):
        valid = ("index.html", b"index", tarfile.REGTYPE)
        for invalid in ("../outside", "/outside", "assets/../../outside", "assets\\outside", ".git/config"):
            with self.subTest(path=invalid), self.assertRaises(ValueError):
                read_bundle(self.archive([valid, (invalid, b"x", tarfile.REGTYPE)]))
        for kind in (tarfile.SYMTYPE, tarfile.LNKTYPE):
            with self.subTest(kind=kind), self.assertRaises(ValueError):
                read_bundle(self.archive([valid, ("assets/link", b"", kind)]))
        with self.assertRaises(ValueError):
            read_bundle(self.archive([valid, valid]))
        with self.assertRaises(ValueError):
            read_bundle(self.archive([("assets/only.js", b"js", tarfile.REGTYPE)]))


if __name__ == "__main__":
    unittest.main()
