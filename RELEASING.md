# Creating a release

Releases are created from Git tags. The tag is the project version; there is no
separate version file. Both Go code changes and changes under `data/notes` require
a new release before they appear on the live site.

## Release steps

1. Commit the changes and verify the project:

    ```sh
    go test ./...
    git diff --check
    git status
    ```

2. Push the release commit to `main`:

    ```sh
    git push origin main
    ```

3. Create and push an annotated semantic-version tag. Use a patch release for
   content updates and fixes, a minor release for new features, and a major
   release for breaking changes.

    ```sh
    git tag -a v1.0.0 -m "Release v1.0.0"
    git push origin v1.0.0
    ```

4. Follow the **Release and deploy Go app** workflow in GitHub Actions and check
   the new entry on the [GitHub Releases page](https://github.com/zanyoats/cconroy.com/releases).

The workflow tests the project, builds a static Linux/AMD64 binary compatible
with Alpine, publishes the binary and its checksum, and deploys that exact tag.
The server sparse-checks out `data/notes` from the same tag, restarts the OpenRC
service, runs a health check, and rolls back the application if the check fails.

## Verify production

```sh
curl -fsS https://cconroy.com/ >/dev/null
curl -fsS https://cconroy.com/feed.xml >/dev/null
ssh blogSite 'sudo rc-service cconroy status'
```

If deployment fails for a transient reason, use **Re-run failed jobs**. If the
tagged code or workflow needs a correction, commit the fix and create a new
version instead of moving or replacing a published tag.
