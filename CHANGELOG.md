# Change Log

## v0.10.0 (2025-11-27)
### Features
* Added helm
* Added node
### Chore
* Updated dependencies

## v0.9.0 (2025-08-27)
### Features
* Do not publish changes again if the same branch exists with the same changes
* Added linux ARM64 build

## v0.8.1 (2025-04-30)
### Features
* Added kubernetes manager
### Fixes
* Fix platform defaults
* Set the git committer instead of the author

## v0.8.0 (2025-04-29)
### Features
* Added Gitea platform
### Chore
* Updated dependencies

## v0.7.0 (2025-03-31)
### Breaking
* Don't use any extractVersion as default
### Features
* Added major-minor-patch-fixed versioning preset
* Added docker-compose manager
### Chore
* Updated dependencies

## v0.6.9 (2025-02-05)
### Fixes
* Only add GitHub token if it is correctly set
### Chore
* Updated dependencies

## v0.6.8 (2025-01-21)
### Fixes
* Correctly inherit ClearFilePatterns

## v0.6.7 (2025-01-21)
### Features
* Add ClearFilePatterns to allow clearing default patterns to only use custom ones

## v0.6.6 (2024-12-03)
### Fixes
* Prevent the same file from being added multiple times from different match-patterns

## v0.6.5 (2024-11-15)
### Features
* Small logging improvements

## v0.6.4 (2024-11-07)
### Features
* Added gitlab_packages datasource

## v0.6.3 (2024-10-10)
### Features
* Small adjustments for creating managers and datasources
* Handle optional "v" by default for versions

## v0.6.2 (2024-10-04)
### Features
* Added GetDatasource to the config object

## v0.6.1 (2024-10-04)
### Features
* Unified manager creation from config
* Moved platform type into platform config
* Merged host rules
* Updated some dependencies

## v0.6.0 (2024-10-01)
### Features
* Restructured the project so that parts of gonovate can be used as go modules by 3rd parties

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
