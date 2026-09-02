import { execFileSync } from "node:child_process";

const containerName = "jobmatcha-mise-smoke";
const volumeName = "jobmatcha-mise-smoke-data";

function docker(args, options = {}) {
  return execFileSync("docker", args, { encoding: "utf8", ...options });
}

function cleanup() {
  try {
    docker(["rm", "--force", containerName], { stdio: "ignore" });
  } catch {}
  try {
    docker(["volume", "rm", volumeName], { stdio: "ignore" });
  } catch {}
}

try {
  cleanup();
  docker(["build", "--tag", "jobmatcha:smoke", "."], { stdio: "inherit" });
  docker(["volume", "create", volumeName], { stdio: "ignore" });
  docker([
    "run",
    "--detach",
    "--name",
    containerName,
    "--publish",
    "127.0.0.1::8181",
    "--volume",
    `${volumeName}:/data`,
    "jobmatcha:smoke",
  ], { stdio: "ignore" });

  const address = docker(["port", containerName, "8181/tcp"]).trim().split(/\r?\n/, 1)[0];
  const endpoint = `http://${address}`;
  let healthy = false;
  for (let attempt = 0; attempt < 20; attempt += 1) {
    try {
      const health = await fetch(`${endpoint}/api/health`);
      if (health.ok && (await health.text()).includes('"status":"ok"')) {
        const spa = await fetch(`${endpoint}/`);
        if (spa.ok && (await spa.text()).toLowerCase().includes("<!doctype html>")) {
          healthy = true;
          break;
        }
      }
    } catch {}
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }

  if (!healthy) {
    docker(["logs", containerName], { stdio: "inherit" });
    throw new Error("Container did not serve a healthy API and SPA within 20 seconds.");
  }
} finally {
  cleanup();
}
