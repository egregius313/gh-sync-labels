# `gh sync-labels`

```
Synchronize labels from one GitHub repository to one or more target repositories

Usage:
  gh sync-labels [flags]

Flags:
      --from string     Source repository (required)
  -h, --help            help for gh
      --label strings   Labels to sync (can be specified multiple times)
      --to strings      Target repositories (can be specified multiple times)
  -u, --update          Update label if it already exists in target repository
  -v, --verbose         Enable verbose output
```

Copies the definition of the labels from one repository to others.