import { spawn } from "node:child_process";
import net from "node:net";

const pnpm = "pnpm";
const air = "air";
const services = [
  { name: "server", cwd: "server", command: air, args: [] },
  {
    name: "web",
    cwd: "web",
    command: pnpm,
    args: ["exec", "vite", "dev", "--port", "3000", "--strictPort"],
  },
];
const ports = [3000, 4321, 8181];
let stopping = false;
const children = [];

function run(command, args, options) {
  if (process.platform !== "win32") return spawn(command, args, options);
  const commandLine = [command, ...args].join(" ");
  return spawn("cmd.exe", ["/d", "/s", "/c", commandLine], options);
}

function prefix(stream, name) {
  let pending = "";
  stream.on("data", (chunk) => {
    pending += chunk;
    const lines = pending.split(/\r?\n/);
    pending = lines.pop();
    for (const line of lines) console.log(`[${name}] ${line}`);
  });
  stream.on("end", () => {
    if (pending) console.log(`[${name}] ${pending}`);
  });
}

function ensurePortAvailable(port) {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.once("error", () => reject(new Error(`Port ${port} is already in use. Stop the existing service before running dev:all.`)));
    server.listen({ host: "127.0.0.1", port }, () => server.close(resolve));
  });
}

function stopChild(child) {
  if (child.exitCode !== null || child.signalCode !== null) return Promise.resolve();
  if (process.platform !== "win32") {
    child.kill("SIGTERM");
    return Promise.resolve();
  }
  return new Promise((resolve) => {
    const taskkill = spawn("taskkill", ["/pid", String(child.pid), "/t", "/f"], { stdio: "ignore" });
    taskkill.once("close", resolve);
    taskkill.once("error", resolve);
  });
}

async function stop(exitCode) {
  if (stopping) return;
  stopping = true;
  await Promise.all([
    ...children.map(stopChild),
    new Promise((resolve) => {
      const astro = run(pnpm, ["exec", "astro", "dev", "stop"], {
        cwd: "docs",
        stdio: "ignore",
      });
      astro.once("close", resolve);
      astro.once("error", resolve);
    }),
  ]);
  process.exit(exitCode);
}

function runDocsInBackground() {
  return new Promise((resolve, reject) => {
    const child = run(pnpm, ["exec", "astro", "dev", "--background"], {
      cwd: "docs",
      stdio: ["inherit", "pipe", "pipe"],
    });
    prefix(child.stdout, "docs");
    prefix(child.stderr, "docs");
    child.once("error", reject);
    child.once("close", (code) => (code === 0 ? resolve() : reject(new Error(`Docs server failed with exit code ${code}.`))));
  });
}

for (const port of ports) await ensurePortAvailable(port);
await runDocsInBackground();

for (const service of services) {
  const child = run(service.command, service.args, {
    cwd: service.cwd,
    stdio: ["inherit", "pipe", "pipe"],
  });
  children.push(child);
  prefix(child.stdout, service.name);
  prefix(child.stderr, service.name);
  child.once("error", (error) => {
    console.error(`[${service.name}] ${error.message}`);
    void stop(1);
  });
  child.once("close", (code, signal) => {
    if (!stopping) {
      console.error(`[${service.name}] stopped unexpectedly (${signal ?? `exit code ${code}`}).`);
      void stop(code ?? 1);
    }
  });
}

const docsLogs = run(pnpm, ["exec", "astro", "dev", "logs", "--follow"], {
  cwd: "docs",
  stdio: ["inherit", "pipe", "pipe"],
});
children.push(docsLogs);
prefix(docsLogs.stdout, "docs");
prefix(docsLogs.stderr, "docs");
docsLogs.once("error", (error) => {
  console.error(`[docs] ${error.message}`);
  void stop(1);
});
docsLogs.once("close", (code, signal) => {
  if (!stopping) {
    console.error(`[docs] stopped unexpectedly (${signal ?? `exit code ${code}`}).`);
    void stop(code ?? 1);
  }
});

process.once("SIGINT", () => void stop(0));
process.once("SIGTERM", () => void stop(0));
