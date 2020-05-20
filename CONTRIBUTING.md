### Contributing

Pull requests for bug fixes are welcome, but before submitting new features or changes to current functionalities [open an issue](https://github.com/DataDog/dd-trace-go/issues/new)
and discuss your ideas or propose the changes you wish to make. After a resolution is reached a PR can be submitted for review.

For commit messages, try to use the same conventions as most Go projects, for example:
```
contrib/database/sql: use method context on QueryContext and ExecContext

QueryContext and ExecContext were using the wrong context to create
spans. Instead of using the method's argument they were using the
Prepare context, which was wrong.

Fixes #113
```
Please apply the same logic for Pull Requests, start with the package name, followed by a colon and a description of the change, just like
the official [Go language](https://github.com/golang/go/pulls).


### Releasing a new version

1. Bump version number in `version.go`.
2. Commit the change and if applicable, add release notes as the commit message.
3. Tag the commit with the new version number prefixed by `v`. For example, if version.go container `Version = 1.2.3` then the git tag should be `v1.2.3`.
4. Push the commit and the tag to Github. At this point the library will be published and can be downloaded by users.
5. To document the new release, create a new release on Github for the newly pushed tag.