import assert from "node:assert/strict";
import test from "node:test";

import { selectJustDeployedVersion } from "./select_deployed_version.mjs";

const before = [{ id: "old-first" }, { id: "old-second" }];
const after = [{ id: "old-first" }, { id: "new-version" }, { id: "old-second" }];

test("selects only latest deployment's new 100% version, never versions list ordering", () => {
  assert.equal(selectJustDeployedVersion({
    before,
    after,
    deployment: { versions: [{ version_id: "new-version", percentage: 100 }] },
  }), "new-version");
});

test("rejects a split deployment", () => {
  assert.throws(() => selectJustDeployedVersion({
    before,
    after,
    deployment: { versions: [{ version_id: "new-version", percentage: 50 }, { version_id: "old-first", percentage: 50 }] },
  }), /exactly one version at 100%/);
});

test("rejects a deployment that did not create a new version", () => {
  assert.throws(() => selectJustDeployedVersion({
    before,
    after,
    deployment: { versions: [{ version_id: "old-first", percentage: 100 }] },
  }), /already present before deploy/);
});
