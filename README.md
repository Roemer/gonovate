# gonovate
Update your dependencies with Go.

## Introduction
gonovate is a tool that allows updating dependencies of your projects.
One of the main goals of gonovate is developer-friendlyness, so it should be easy to configure, test and use.

## Concepts
Most components from gonovate are similar to the ones from other updaters.
The main components are:
- Configuration: One or more json files which hold the configuration
- Preset: One or more configuration files which can be used as a base of your configuration
- Platform: Defines how the project to check for updates is checked out and how to inform about updates
- Manager: Searches for existing dependencies and knows how to update them
- Datasource: Searches for updates for existing dependencies
- Dependency: A concrete dependency that has a version and might need updating
- Rules: A very flexible way to configure all parts of gonovate
- HostRules: Contains credentials to access secured datasources to check for updates

## Configuration
There is usually a `gonovate.json` file which contains your configuration. The basic structure of this file is:

```json
{
    "platform": {
        "type": "noop",
        ...
    },
    "extends": [
        "defaults"
    ],
    "managers": [
        ...
    ],
    "rules": [
        ...
    ],
    "hostRules": [
        ...
    ]
}
```

## Platforms
Platforms are the component that interact with your project. That means, it is responsible for cloning, creating branches, pushing changes and creating pull-requests.

The following platforms are available:
| platform | description |
| --- | --- |
| git | This platform just uses git features. So it cannot create pull-requests for example. |
| github | This platform supports all features and interacts with projects hosted on GitHub. |
| gitlab | This platform supports all features and interacts with projects hosted on GitLab.|
| noop | This platform does not implement any features. It is most usefull for applying changes locally without branches or commits. |

## Managers
Managers are the components that are responsible for finding dependencies in your project and writing back updates.

The following managers are available:
| manager | description |
| --- | --- |
| devcontainer | This manager updates devcontainer.json files. |
| dockerfile | This manager updates Dockerfiles. |
| gomod | This manager updates Go dependencies. |
| inline | This manager uses inline comments in files to search dependencies in those files. |
| regex | This manager uses regular expressions to search for dependencies. |

### Manager Configuration
A manager needs an `id` and a `type` and contains `managerConfig` which configure which files should be handled by the manager and how.
Additionally, it can contain `dependencyConfig` that define some behavior for all dependencies that are handled by this manager.

Example:
```json
{
    "id": "dockerfile-versions",
    "type": "regex",
    "managerConfig": {
        "filePatterns": [
            "**/[Dd]ockerfile"
        ],
        "matchStrings": [
            "^ENV .*?_VERSION=(?P<version>.*) # (?P<datasource>.*?)\/(?P<dependencyName>.*?)[[:blank:]]*$"
        ]
    },
    "dependencyConfig": {
        "maxUpdateType": "major"
    }
}
```

## Datasources
Datasources are responsible for fetching available versions for the dependencies.
With that information, gonovate can decide which version a dependency should update to if there is an update.

The following datasources are available:
| manager | description |
| --- | --- |
| artifactory | Fetches information from a self hosted artifactory. |
| docker | Fetches information from any docker registry. |
| github_releases | Fetches information from GitHub releases. |
| github_tags | Fetches information from GitHub tags. |
| go_mod | Fetches information for go modules. |
| go_version | Fetches information for the go version. |
| gradle_version | Fetches information for the gradle version. |
| java_version | Fetches information for the java version. |
| maven | Fetches information for maven modules. |
| nodejs | Fetches information for the node version. |
| npm | Fetches information for npm modules. |

## Rules
Rules allow customizing managers and the handling of dependencies in a flexible way.

### Rules Configuration
The rules contain a `matches` section which describe the criteria when this rule should match and `managerConfig` and/or `dependencyConfig` that should be applied when this rule matches.

Example:
```json
{
    "matches": {
        "managers": [
            "dockerfile-versions"
        ]
    },
    "managerConfig": {
        "filePatterns": [
            "**/my.[Dd]ockerfile"
        ]
    }
}
```

## Host Rules
Host rules contain credentials that might be needed when accessing datasources to check for newer versions.

Example: 
```json
{
    "matchHost": "index.docker.io",
    "username": "docker_user",
    "password": "dckr_pat_abcdefghijklmnop"
}
```
