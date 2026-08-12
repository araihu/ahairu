import {
  argument,
  dag,
  Container,
  Directory,
  File,
  func,
  object,
  Secret,
} from "@dagger.io/dagger"

const NODE_IMAGE =
  "node:24-bookworm@sha256:934240a162082fd8b8a2f90cd5114446443f1eba1c5378f6687167ca405e6584"
const GO_IMAGE =
  "golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd"
const PYTHON_IMAGE =
  "python:3.14-bookworm@sha256:8771427e2ac3e39208c1632f17e8b09e464333d262844a03705cc5e0023c16e2"

const SOURCE_EXCLUDES = [
  ".git",
  ".git/**",
  ".dagger/node_modules",
  ".dagger/node_modules/**",
  "node_modules",
  "node_modules/**",
  "public",
  "public/**",
  ".wrangler",
  ".wrangler/**",
  "__pycache__",
  "**/__pycache__/**",
]

@object()
export class Ahairu {
  /** Run ordinary pull-request/source validation without protected Assets credentials. */
  @func({ cache: "never" })
  async source(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    runNonce: string,
  ): Promise<string> {
    return this.sourceContainer(source)
      .withEnvVariable("AHAIRU_RUN_NONCE", runNonce)
      .withExec(["npm", "audit", "--audit-level=high"])
      .withExec(["templ", "generate"])
      .withExec([
        "git",
        "diff",
        "--no-index",
        "--exit-code",
        "--",
        "/src/site",
        "/work/site",
      ])
      .withExec(["go", "test", "./...", "-count=1"])
      .withExec(["npm", "run", "test:workflow"])
      .withExec(["npm", "run", "test:routes"])
      .withExec(["npm", "run", "test:canary"])
      .stdout()
  }

  /** Materialize a dispatch handoff, run the complete safe gate, and return normalized handoff JSON. */
  @func({ cache: "never" })
  async acceptedAssetsDispatch(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    handoffJson: Secret,
    assetsGithubToken: Secret,
    runNonce: string,
    dispatchEventType = "araihu-assets-released",
  ): Promise<File> {
    if (dispatchEventType !== "araihu-assets-released") {
      throw new Error(`unexpected dispatch event type: ${dispatchEventType}`)
    }
    return this.acceptedAssets(source, assetsGithubToken)
      .withEnvVariable("AHAIRU_RUN_NONCE", runNonce)
      .withExec(["npm", "audit", "--audit-level=high"])
      .withSecretVariable("ASSETS_HANDOFF_JSON", handoffJson)
      .withExec([
        "bash",
        "-euo",
        "pipefail",
        "-c",
        'python3 scripts/prepare_asset_bundle.py --handoff-json "$ASSETS_HANDOFF_JSON" --accepted-output /tmp/accepted-assets.json --output /tmp/asset-bundle',
      ])
      .withEnvVariable("ASSET_BUNDLE", "/tmp/asset-bundle")
      .withExec(["npm", "run", "check"])
      .file("/tmp/accepted-assets.json")
  }

  /** Materialize a main-promotion handoff, run the complete safe gate, and return normalized handoff JSON. */
  @func({ cache: "never" })
  acceptedAssetsMain(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    handoffJson: Secret,
    assetsGithubToken: Secret,
    runNonce: string,
  ): File {
    return this.acceptedAssets(source, assetsGithubToken)
      .withEnvVariable("AHAIRU_RUN_NONCE", runNonce)
      .withExec(["npm", "audit", "--audit-level=high"])
      .withSecretVariable("ASSETS_HANDOFF_JSON", handoffJson)
      .withExec([
        "bash",
        "-euo",
        "pipefail",
        "-c",
        'python3 scripts/prepare_asset_bundle.py --handoff-json "$ASSETS_HANDOFF_JSON" --accepted-output /tmp/accepted-assets.json --output /tmp/asset-bundle',
      ])
      .withEnvVariable("ASSET_BUNDLE", "/tmp/asset-bundle")
      .withExec(["npm", "run", "check"])
      .file("/tmp/accepted-assets.json")
  }

  /** Reacquire an accepted immutable bundle, run the same full gate, then promote Cloudflare. */
  @func({ cache: "never" })
  deploy(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    acceptedAssets: File,
    assetsGithubToken: Secret,
    cloudflareApiToken: Secret,
    cloudflareAccountId: Secret,
    runNonce: string,
  ): File {
    const container = this.projectContainer(source)
      .withFile("/tmp/accepted-assets.json", acceptedAssets)
      .withSecretVariable("ASSETS_GITHUB_TOKEN", assetsGithubToken)
      .withEnvVariable("AHAIRU_RUN_NONCE", runNonce)
      .withExec([
        "python3",
        "scripts/prepare_asset_bundle.py",
        "--handoff-file",
        "/tmp/accepted-assets.json",
        "--output",
        "/tmp/asset-bundle",
      ])
      .withEnvVariable("ASSET_BUNDLE", "/tmp/asset-bundle")
      .withExec(["npm", "run", "check"])
      .withSecretVariable("CLOUDFLARE_API_TOKEN", cloudflareApiToken)
      .withSecretVariable("CLOUDFLARE_ACCOUNT_ID", cloudflareAccountId)
      .withEnvVariable("WRANGLER_OUTPUT_FILE_PATH", "/tmp/wrangler-version-upload.jsonl")
      .withExec(["npm", "exec", "--", "wrangler", "versions", "upload"])
      .withExec([
        "node",
        "scripts/select_deployed_version.mjs",
        "--upload",
        "/tmp/wrangler-version-upload.jsonl",
      ], { redirectStdout: "/tmp/uploaded-version" })
      .withExec([
        "bash",
        "-euo",
        "pipefail",
        "-c",
        "npm exec -- wrangler versions deploy --version-id \"$(tr -d '\\n' </tmp/uploaded-version)\" --percentage 100 --yes",
      ])
      .withExec([
        "npm",
        "exec",
        "--",
        "wrangler",
        "deployments",
        "status",
        "--json",
      ], { redirectStdout: "/tmp/worker-deployment.json" })
      .withEnvVariable("GITHUB_STEP_SUMMARY", "/dev/null")
      .withExec([
        "bash",
        "-euo",
        "pipefail",
        "-c",
        "node scripts/select_deployed_version.mjs --uploaded-version \"$(tr -d '\\n' </tmp/uploaded-version)\" --deployment /tmp/worker-deployment.json",
      ])
    return container.file("/tmp/uploaded-version")
  }

  /** Persist post-promotion accepted Assets state on the dedicated automation ref. */
  @func({ cache: "never" })
  async acceptDeployedAssetsState(
    @argument({ defaultPath: ".", ignore: SOURCE_EXCLUDES }) source: Directory,
    acceptedAssets: File,
    ahairuGithubToken: Secret,
    stateBaseSha: string,
    runNonce: string,
    githubApiUrl = "https://api.github.com",
  ): Promise<string> {
    const script = String.raw`set -euo pipefail
python3 scripts/accepted_assets_state.py --handoff-file /tmp/accepted-assets.json --output /tmp/accepted-channel-v1.json
headers=(
  --header "Accept: application/vnd.github+json"
  --header "Authorization: Bearer \${GH_TOKEN}"
  --header "X-GitHub-Api-Version: 2022-11-28"
)
ref_url="\${GITHUB_API_URL}/repos/\${AHAIRU_REPOSITORY}/git/ref/heads/\${STATE_REF}"
ref_response=/tmp/accepted-state-ref.json
ref_status=$(curl --silent --show-error --output "$ref_response" --write-out '%{http_code}' "\${headers[@]}" "$ref_url")
if [[ "$ref_status" == 404 ]]; then
  python3 - /tmp/accepted-state-ref-create.json "$STATE_REF" "$STATE_BASE_SHA" <<'PY'
import json, sys
path, ref, sha = sys.argv[1:]
with open(path, "x", encoding="utf-8") as output:
    json.dump({"ref": "refs/heads/" + ref, "sha": sha}, output, separators=(",", ":"))
PY
  create_status=$(curl --silent --show-error --output /tmp/accepted-state-ref-created.json --write-out '%{http_code}' \
    --request POST "\${headers[@]}" --data-binary @/tmp/accepted-state-ref-create.json "\${GITHUB_API_URL}/repos/\${AHAIRU_REPOSITORY}/git/refs")
  if [[ "$create_status" == 422 ]]; then
    ref_status=$(curl --silent --show-error --output "$ref_response" --write-out '%{http_code}' "\${headers[@]}" "$ref_url")
  elif [[ "$create_status" != 201 ]]; then
    echo "Create dedicated accepted-state ref failed (HTTP $create_status)" >&2
    exit 1
  else
    ref_status=200
  fi
fi
test "$ref_status" = 200 || { echo "Dedicated accepted-state ref is unavailable (HTTP $ref_status)" >&2; exit 1; }

state_status=$(curl --silent --show-error --output /tmp/accepted-state-existing.json --write-out '%{http_code}' "\${headers[@]}" \
  "\${GITHUB_API_URL}/repos/\${AHAIRU_REPOSITORY}/contents/\${STATE_PATH}?ref=\${STATE_REF}")
state_sha=""
case "$state_status" in
  200) state_sha=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1], encoding="utf-8"))["sha"])' /tmp/accepted-state-existing.json) ;;
  404) ;;
  *) echo "Read accepted state failed (HTTP $state_status)" >&2; exit 1 ;;
esac
python3 - /tmp/accepted-state-update.json "$STATE_REF" "$state_sha" /tmp/accepted-channel-v1.json <<'PY'
import base64, json, pathlib, sys
output, ref, sha, state_path = sys.argv[1:]
payload = {
    "message": "automation: accept deployed Assets channel",
    "content": base64.b64encode(pathlib.Path(state_path).read_bytes()).decode("ascii"),
    "branch": ref,
}
if sha:
    payload["sha"] = sha
with open(output, "x", encoding="utf-8") as stream:
    json.dump(payload, stream, separators=(",", ":"))
PY
update_status=$(curl --silent --show-error --output /tmp/accepted-state-updated.json --write-out '%{http_code}' \
  --request PUT "\${headers[@]}" --data-binary @/tmp/accepted-state-update.json \
  "\${GITHUB_API_URL}/repos/\${AHAIRU_REPOSITORY}/contents/\${STATE_PATH}")
if [[ -z "$state_sha" ]]; then
  test "$update_status" = 201 || { echo "Create accepted state conflicted or failed (HTTP $update_status)" >&2; exit 1; }
else
  test "$update_status" = 200 || { echo "Update accepted state conflicted or failed (HTTP $update_status)" >&2; exit 1; }
fi
echo "accepted Assets state updated on \${STATE_REF}"`.replaceAll("\\${", "${")

    return dag
      .container()
      .from(PYTHON_IMAGE)
      .withEnvVariable("AHAIRU_RUN_NONCE", runNonce)
      .withExec(["apt-get", "update"])
      .withExec(["apt-get", "install", "-y", "--no-install-recommends", "ca-certificates", "curl"])
      .withExec(["rm", "-rf", "/var/lib/apt/lists/*"])
      .withDirectory("/work", source)
      .withWorkdir("/work")
      .withFile("/tmp/accepted-assets.json", acceptedAssets)
      .withSecretVariable("GH_TOKEN", ahairuGithubToken)
      .withEnvVariable("GITHUB_API_URL", githubApiUrl)
      .withEnvVariable("AHAIRU_REPOSITORY", "araihu/ahairu")
      .withEnvVariable("STATE_REF", "automation/araihu-assets-state")
      .withEnvVariable("STATE_PATH", ".automation/araihu-assets/accepted-channel-v1.json")
      .withEnvVariable("STATE_BASE_SHA", stateBaseSha)
      .withExec(["bash", "-c", script])
      .stdout()
  }

  private sourceContainer(source: Directory): Container {
    return this.projectContainer(source)
  }

  private acceptedAssets(source: Directory, assetsGithubToken: Secret): Container {
    return this.projectContainer(source)
      .withSecretVariable("ASSETS_GITHUB_TOKEN", assetsGithubToken)
  }

  private projectContainer(source: Directory): Container {
    const npmCache = dag.cacheVolume("ahairu-npm-v1")
    const goBuildCache = dag.cacheVolume("ahairu-go-build-v1")
    const goModuleCache = dag.cacheVolume("ahairu-go-mod-v1")
    const puppeteerCache = dag.cacheVolume("ahairu-puppeteer-v1")
    const templBuildCache = dag.cacheVolume("ahairu-templ-go-build-v1")
    const templModuleCache = dag.cacheVolume("ahairu-templ-go-mod-v1")
    const goDistribution = dag.container().from(GO_IMAGE).directory("/usr/local/go")
    const templ = dag
      .container()
      .from(GO_IMAGE)
      .withMountedCache("/go/pkg/mod", templModuleCache)
      .withMountedCache("/root/.cache/go-build", templBuildCache)
      .withExec(["go", "install", "github.com/a-h/templ/cmd/templ@v0.3.1020"])
      .file("/go/bin/templ")

    return dag
      .container()
      .from(NODE_IMAGE)
      .withExec(["apt-get", "update"])
      .withExec([
        "apt-get",
        "install",
        "-y",
        "--no-install-recommends",
        "bash",
        "ca-certificates",
        "curl",
        "fonts-liberation",
        "git",
        "libasound2",
        "libatk-bridge2.0-0",
        "libatk1.0-0",
        "libcups2",
        "libdbus-1-3",
        "libdrm2",
        "libgbm1",
        "libgtk-3-0",
        "libnspr4",
        "libnss3",
        "libx11-xcb1",
        "libxcomposite1",
        "libxdamage1",
        "libxfixes3",
        "libxkbcommon0",
        "libxrandr2",
        "python3",
        "xdg-utils",
      ])
      .withExec(["rm", "-rf", "/var/lib/apt/lists/*"])
      .withDirectory("/src", source)
      .withDirectory("/work", source, { owner: "node:node" })
      .withDirectory("/usr/local/go", goDistribution)
      .withWorkdir("/work")
      .withFile("/usr/local/bin/templ", templ, { permissions: 0o755 })
      .withMountedCache("/home/node/.npm", npmCache, { owner: "node:node" })
      .withMountedCache("/home/node/.cache/go-build", goBuildCache, { owner: "node:node" })
      .withMountedCache("/home/node/go/pkg/mod", goModuleCache, { owner: "node:node" })
      .withMountedCache("/home/node/.cache/puppeteer", puppeteerCache, { owner: "node:node" })
      .withEnvVariable("HOME", "/home/node")
      .withEnvVariable("GOCACHE", "/home/node/.cache/go-build")
      .withEnvVariable("GOPATH", "/home/node/go")
      .withEnvVariable("PATH", "/usr/local/go/bin:/home/node/go/bin:$PATH", { expand: true })
      .withUser("node")
      .withExec(["npm", "ci"])
  }
}
