import {Octokit} from "@octokit/action";
import * as fs from 'fs';
import * as path from 'path';

let distributions = [];
const dir = path.join(path.resolve(), 'distributions');

for (const file of fs.readdirSync(dir)) {
  if (fs.existsSync(path.join(dir, file, 'manifest.yaml'))) {
    distributions.push(file);
  }
}

if (distributions.length === 0) {
  console.error("No distributions found!");
  process.exit(1);
}

console.log("Distributions found:", distributions);

const octokit = new Octokit();
const [owner, repo] = process.env.GITHUB_REPOSITORY.split("/");

const response = await octokit.request("GET /repos/{owner}/{repo}/releases", {
  owner,
  repo,
  per_page: 50,
  headers: {
    'X-GitHub-Api-Version': '2022-11-28'
  }
});

// Generate regex from distributions to match the asset name and parts
// Example: {nrdot-collector-host}_{1.0.0}_{linux}_{amd64}.{tar.gz}
const regex = new RegExp(`^(${distributions.join("|")})_([0-9]+\.[0-9]+\.[0-9]+)_([a-z]+)_([a-z0-9_]+)\.([a-z0-9\.]+)$`);

const currentTime = new Date().getTime();

const metrics = [];

for (const release of response.data) {
  console.log("Processing release:", release.tag_name);

  for (const asset of release.assets) {

    if (asset.name.endsWith('.asc') || asset.name.endsWith('.sum')) {
      console.log("Skipping signature or checksum file:", asset.name);
      continue;
    }

    const match = asset.name.match(regex);
    if (match) {
      console.log("Matched asset:", asset.name);

      const [distribution, version, os, arch, ext] = match.slice(1);

      metrics.push({
        name: "nrdot.collector.downloads.package",
        type: "gauge",
        value: asset.download_count,
        attributes: {
          "nrdot.distro": distribution,
          "nrdot.version": version,
          "package.os": os,
          "package.arch": arch,
          "package.ext": ext
        },
        timestamp: currentTime
      });

    } else {
      console.log("Ignoring asset:", asset.name);
    }
  }
}

// Send the payload to New Relic if there are any metrics
if (metrics.length === 0) {
  console.error("No metrics to send.");
  process.exit(1);
}

// Send the payload to New Relic
console.log("Metrics to be sent:", metrics.length);

const options = {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Api-Key': process.env.NEW_RELIC_LICENSE_KEY
  },
  body: JSON.stringify([{metrics}])
};

const nrResponse = await fetch("https://metric-api.newrelic.com/metric/v1", options);

if (!nrResponse.ok) {
  console.error("Error sending metrics:", nrResponse.status, nrResponse.statusText);
  process.exit(1);
}

console.log("Metrics sent successfully:", nrResponse.status, nrResponse.statusText);
