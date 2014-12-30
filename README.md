Misc tools for working with git and Gerrit.

For now only `git-cleanup`.

Install with:

```
  $ go get github.com/bradfitz/gitutil/git-cleanup
```

Use like:

```
  $ git checkout master
  $ git sync
  $ git cleanup
```

(Assuming you're contributing to Go with `git-codereview` and Gerrit, ala https://golang.org/doc/contribute.html)

