import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const ROOT = resolve(import.meta.dirname, "..").replaceAll("\\", "/");
const CORE_MINIMUM = 80;
const OVERALL_MINIMUM = 70;

function normalize(path) {
	const normalized = path.replaceAll("\\", "/");
	if (normalized.startsWith(ROOT)) return normalized.slice(ROOT.length);
	const modulePath = "github.com/caslus/jobmatcha/";
	if (normalized.startsWith(modulePath)) return `/server/${normalized.slice(modulePath.length)}`;
	return normalized;
}

function excluded(path) {
	return [
		/^\/server\/(cmd|migrations)\//,
		/^\/server\/internal\/(scanner\/providers|testutil)\//,
		/^\/web\/(dist|coverage)\//,
		/^\/web\/src\/(test|types\/api\.gen\.ts|routeTree\.gen\.ts|router\.tsx|integrations\/tanstack-query\/devtools\.tsx)/,
		/\.test\.[jt]sx?$/,
	].some((pattern) => pattern.test(path));
}

function core(path) {
	return /^\/server\/internal\/(api|app|repository|service|util)\//.test(path);
}

function empty() {
	return { covered: 0, total: 0 };
}

function add(target, covered, total) {
	target.covered += covered;
	target.total += total;
}

function parseGoProfile(profile) {
	const overall = empty();
	const coreCoverage = empty();
	const blocks = new Map();
	for (const line of profile.trim().split(/\r?\n/).slice(1)) {
		const match = /^(.*?):(\d+\.\d+,\d+\.\d+)\s+(\d+)\s+(\d+)$/.exec(line);
		if (!match) continue;
		const path = normalize(match[1]);
		if (excluded(path)) continue;
		const statements = Number(match[3]);
		const hits = Number(match[4]);
		const key = `${path}:${match[2]}:${statements}`;
		blocks.set(key, { path, statements, hits: Math.max(hits, blocks.get(key)?.hits ?? 0) });
	}
	for (const { path, statements, hits } of blocks.values()) {
		add(overall, hits > 0 ? statements : 0, statements);
		if (core(path)) add(coreCoverage, hits > 0 ? statements : 0, statements);
	}
	return { overall, core: coreCoverage };
}

function parseLcov(report) {
	const overall = empty();
	for (const record of report.split("end_of_record")) {
		const source = record.match(/^SF:(.+)$/m)?.[1];
		if (!source) continue;
		const path = normalize(resolve(source));
		if (excluded(path)) continue;
		for (const line of record.matchAll(/^DA:\d+,(-?\d+)$/gm)) {
			add(overall, Number(line[1]) > 0 ? 1 : 0, 1);
		}
	}
	return overall;
}

function percentage({ covered, total }) {
	return total === 0 ? 0 : (covered / total) * 100;
}

const [goProfile, webLcov] = await Promise.all([
	readFile(resolve(ROOT, "server/coverage.out"), "utf8"),
	readFile(resolve(ROOT, "web/coverage/lcov.info"), "utf8"),
]);
const go = parseGoProfile(goProfile);
const web = parseLcov(webLcov);
const overall = { ...go.overall };
add(overall, web.covered, web.total);

const results = [
	["core backend application code", go.core, CORE_MINIMUM],
	["overall maintained application code", overall, OVERALL_MINIMUM],
];

let failed = false;
for (const [name, coverage, minimum] of results) {
	const value = percentage(coverage);
	console.log(`${name}: ${value.toFixed(2)}% (${coverage.covered}/${coverage.total} weighted lines; minimum ${minimum}%)`);
	if (value < minimum) {
		console.error(`Coverage gate failed: ${name} is below ${minimum}%.`);
		failed = true;
	}
}

if (failed) process.exitCode = 1;
