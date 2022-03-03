This guide assumes your remotes are setup as shown below with the fork as `origin` and the upstream as `upstream`

```
origin  git@github.com:zendesk/scheduled-scaler.git (fetch)
origin  git@github.com:zendesk/scheduled-scaler.git (push)
upstream        git@github.com:West-Coast-Devops/scheduled-scaler.git (fetch)
upstream        git@github.com:West-Coast-Devops/scheduled-scaler.git (push)
```

## Contributing upstream

Run the following commands to have git setup.

1. `git remote add upstream git@github.com:West-Coast-Devops/scheduled-scaler.git`
2. `git fetch upstream`
3. `git checkout -b ${your-branch-name-here} upstream/master`
4. make your changes
5. `git push origin ${your-branch-name-here}`
6. submit a pull request

## Internal stuff

1. `git checkout -b ${your-branch-name-here} origin/master`
2. make changes
3. `git push origin ${your-branch-name-here}`
4. submit a pull request