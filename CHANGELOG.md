# Change Log

## v0.5.1 (2024-09-03)

### Features

* Allow 're:' as prefix for regexes for matching manager ids or depdendency names
* Allow passing the --config parameter multiple times to load multiple configs

### Fixes

* Small check for errors in the cli exclusive flag

## v0.5.0 (2024-08-28)

### Features

* Added support for Docker digests
* Added cli exclusive flag

### Fixes

* Fixed +incompatible in go module versions

## v0.4.0 (2024-08-23)

### Features

* Improved config loader for local files and presets
* Config loader can now load from an url
* Support yaml configs
* Added possibility to skip individual dependencies
* Implemented api tokens for GitHub releases/tags
* Added a finished message
* Improved artifactory authentication (apikey, token, username/password)
* **manager:** Added Devcontainer manager
* **datasource:** Added Java-Version datasource
* **datasource:** Added Gradle-Version datasource
* **datasource:** Added Browser-Version datasource
* **datasource:** Added Ant-Version datasource

### Fixes

* Fixed regex to better parse Dockerfile 'FROM' lines
* Ignore Docker dependencies with 'latest' tag

## v0.3.0 (2024-08-16)

### Features

* Major refactoring for the whole dependency handling
* Implemented grouping
* MRs/PRs now have a better title and a body text
* Implemented platform cleanup (to cleanup stale branches and MRs/PRs)
* Added Go-Mod manager
* Added Dockerfile manager

### Fixes

* Ignore Docker "latest" tag

## v0.2.3 (2024-06-24)

### Fixes

* GitHub tags/releases now use paging

## v0.2.2 (2024-06-06)

### Fixes

* Fix maxUpdateType and extractVersion from inline definition

## v0.2.1 (2024-06-06)

### Features

* Add maxUpdateType and extractVersion to inline definition
* Print version on cli help

## v0.2.0 (2024-06-03)

### Features

* Added branchPrefix configuration
* Added VersioningPresets, add config to datasources
* Throw error when a preset is not found
* When using direct, passing a project is optional

## v0.1.0 (2024-05-29)

### Features

* **manager:** Added Inline Manager
* **manager:** Added Regex Manager
* **platform:** Added git platform
* **platform:** Added GitHub platform
* **platform:** Added GitLab platform
* **datasource:** Added Artifactory datasource
* **datasource:** Added Docker datasource
* **datasource:** Added GitHub-Releases datasource
* **datasource:** Added GitHub-Tags datasource
* **datasource:** Added Go-Version datasource
* **datasource:** Added Maven datasource
* **datasource:** Added NodeJS datasource
* **datasource:** Added NPM datasource
