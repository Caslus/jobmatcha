module.exports = {
  branches: ["main"],
  tagFormat: "v${version}",
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      {
        publishCmd: "bash scripts/publish-container.sh ${nextRelease.version}",
      },
    ],
    "@semantic-release/github",
  ],
};
