# Scripts

Scripts are organized into the following directories:
* **build:** Scripts used in CI to build the collector. All files in this directory are included in the source cache key hash. Only add scripts here if changes to them should invalidate the cache and force a fresh build.
* **release:** Scripts used to facilitate the release process.
* **misc:** Everything else.