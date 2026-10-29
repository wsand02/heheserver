import { execFileSync } from "node:child_process";
const changes = execFileSync(
  "git",
  ["status", "--porcelain", "--", "internal/frontend/dist"],
  { cwd: new URL("../../", import.meta.url), encoding: "utf8" },
);
if (changes.trim()) {
  console.error(
    "Rebuild with pnpm build and commit internal/frontend/dist:\n" + changes,
  );
  process.exit(1);
}
