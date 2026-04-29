<!--
Thanks for the PR! A quick checklist before you submit:
-->

## Summary

<!-- One paragraph: what changed and why. -->

## Type

- [ ] New cleanup rule (no code change)
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation
- [ ] Refactor / internal change

## Checklist

- [ ] `go vet ./...` is clean
- [ ] `go test ./...` passes
- [ ] If I added a rule, I've pasted the relevant `./reclaim --plain` output below
- [ ] If I changed user-facing behavior, I've updated `README.md` and/or `CHANGELOG.md`
- [ ] If I touched safety logic, I've added a test that proves it rejects unsafe input

## Verification output (paste below)

```
$ ./reclaim --plain
…
```
