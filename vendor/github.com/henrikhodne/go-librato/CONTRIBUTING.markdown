# How to contribute

I'd love to accept your patches and contributions to this project. There are
a just a few small guidelines you need to follow.

## Submitting a patch

  0. It's generally best to start by opening a new issue describing the bug or
     feature you're intending to fix. Even if you think it's relatively minor,
     it's helpful to know what people are working on. Mention in the initial
     issue that you are planning to work on that bug or feature so that it can
     be assigned to you.

  0. Follow the normal process of [forking][] the project, and setup a new
     branch to work in.  It's important that each group of changes be done in
     separate branches in order to ensure that a pull request only includes the
     commits related to that bug or feature.

  0. Go makes it very simple to ensure properly formatted code, so always run
     `go fmt` on your code before committing it.  You should also run
     [golint][] over your code.  As noted in the [golint readme][], it's not
     strictly necessary that your code be completely "lint-free", but this will
     help you find common style issues.

  0. Any significant changes should almost always be accompanied by tests.  The
     project already has good test coverage, so look at some of the existing
     tests if you're unsure how to go about it.  [gocov][] and [gocov-html][]
     are invaluable tools for seeing which parts of your code aren't being
     exercised by your tests.

  0. Do your best to have [well-formed commit messages][] for each change.
     This provides consistency throughout the project, and ensures that commit
     messages are able to be formatted properly by various git tools.

  0. Finally, push the commits to your fork and submit a [pull request][].

[forking]: https://help.github.com/articles/fork-a-repo
[golint]: https://github.com/golang/lint
[golint readme]: https://github.com/golang/lint/blob/master/README
[gocov]: https://github.com/axw/gocov
[gocov-html]: https://github.com/matm/gocov-html
[well-formed commit messages]: http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html
[squash]: http://git-scm.com/book/en/Git-Tools-Rewriting-History#Squashing-Commits
[pull request]: https://help.github.com/articles/creating-a-pull-request

## Other notes on code organization

Currently, everything is defined in the main `librato` package, with API methods
broken into separate service objects.  These services map directly to how
the [Librato API documentation][] is organized, so use that as your guide for
where to put new methods.

Sub-service (e.g. [Space Charts][]) implementations are split into separate
files based on the APIs they provide. These files are named service_api.go (e.g.
spaces_charts.go) to describe the API to service mappings.

[Librato API documentation]: http://dev.librato.com/v1
[Space Charts]: http://dev.librato.com/v1/get/spaces/:id/charts
