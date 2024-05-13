# gonovate
Go renovate your dependencies with style

## Introduction
gonovate is a tool that allows updating dependencies of your projects. One of the main goals of gonovate is developer-friendlyness, so it should be easy to configure, test and use.

## Concepts
Most components from gonovate are similar to the ones from other updaters. The main components are:
- Configurations: One or more json files which hold the configuration
- Preset: One or more configuration files which can be used as a base of your configuration
- Manager: Defines a set settings on how to find files, versions and how to update them
- Datasource: A source which can be checked for updates
- Platform: A platform defines how the project to check for updates is checked out and how to inform about updates
- Package: A package is a concrete dependency that has a version and might need updating
- Rules: A very flexible way to configure all parts of a manager or versioning
- HostRules: Contains credentials to access datasources to check for updates

## Managers
The follownig managers are currently implemented:

### Regex
This manager allows to use regular expressions to search for package names and versions which are used to find updates.

## Configuration
There is usually a `gonovate.json` file which contains your specific configuration. The basic structure of this file is:
```json
{
    "platform": "github",
    "extends": [
        "defaults"
    ],
    "ignorePatterns": [
        "**/.git"
    ],
    "managers": [
        {
            ...
        }
    ],
    "rules": [
        {
            ...
        }
    ],
    "hostRules": [
        {
            ...
        }
    ]
}
```

### Manager Configuration
A manager needs an `id` and a `type` and contains `managerSettings` which configure which files should be handled by the manager and how.
Additionally, it can contain `packageSettings` that define some behavior for all packages that are handled by this manager.

Example:
```json
{
    "id": "dockerfile-versions",
    "type": "regex",
    "managerSettings": {
        "filePatterns": [
            "**/[Dd]ockerfile"
        ],
        "matchStrings": [
            "^ENV .*?_VERSION=(?P<version>.*) # (?P<datasource>.*?)\/(?P<package>.*?)[[:blank:]]*$"
        ]
    },
    "packageSettings": {
        "maxUpdateType": "major"
    }
}
```

### Rules Configuration
The rules contain a `matches` section which describe the criteria when this rule should match and `managerSettings` and/or `packageSettings` that should be applied when this rule matches.

Example:
```json
{
    "matches": {
        "managers": [
            "dockerfile-versions"
        ]
    },
    "managerSettings": {
        "filePatterns": [
            "**/my.[Dd]ockerfile"
        ]
    }
}
```

### Host Rules Configuration
Host rules contain credentials that might be needed when accessing datasources to check for newer versions.

Example: 
```json
{
    "matchHost": "index.docker.io",
    "username": "roemer",
    "password": "dckr_pat_abcdefghijklmnop"
}
```
