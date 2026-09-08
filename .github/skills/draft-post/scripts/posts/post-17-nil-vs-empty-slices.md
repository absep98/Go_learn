---
title: The One-Line Go Bug That Turns Your API's Empty List into null
subtitle: Part 17 of Learning Go in Public
tags: go, buildinpublic, backend, api
series: Learning Go in Public
---

I have an endpoint on my personal analytics backend that returns failed background jobs — the ones that errored out and need attention. When there were failures, the API returned a clean JSON array. When there were zero failures, clients started getting `null` back instead of an empty array, and every downstream `.map()` and `.length` check broke.

## What I Tried First (the buggy version)

```go
var failedJobs []models.FailedJob

rows, err := db.Query(ctx, "SELECT ... FROM failed_jobs WHERE ...")
// ...
for rows.Next() {
    var job models.FailedJob
    // scan job...
    failedJobs = append(failedJobs, job)
}

return failedJobs
```

This compiles. It works when there are failed jobs. It silently breaks when there aren't.

## Why It Breaks

`var failedJobs []models.FailedJob` declares a **nil slice** — not an empty one. If the query returns zero rows, the `for` loop body never runs, so `append` is never called, and `failedJobs` stays nil all the way to the response.

Go's `encoding/json` package treats nil slices and empty slices differently:
- A nil slice encodes as `null`
- An empty (but non-nil) slice encodes as `[]`

Same real-world meaning — "no items" — but two different values on the wire. Any client expecting an array and calling `.length` or `.map()` on `null` blows up.

## The Fix

```go
failedJobs := make([]models.FailedJob, 0)

rows, err := db.Query(ctx, "SELECT ... FROM failed_jobs WHERE ...")
// ...
for rows.Next() {
    var job models.FailedJob
    // scan job...
    failedJobs = append(failedJobs, job)
}

return failedJobs
```

One line changed. Now zero results always serialize as `[]`, never `null`.

## Why It Works

`make([]models.FailedJob, 0)` allocates a slice with length 0 but a non-nil underlying value. `append` behaves identically whether the slice starts nil or empty, so nothing else in the function needs to change — only the zero-value case at the boundary.

## What I'd Do Differently

- Any slice or map that crosses an API boundary should be initialized explicitly — don't rely on Go's zero value when a client will parse the result.
- This one is easy to miss because it only shows up when the "happy path" data is empty — exactly the case least likely to get tested manually.
- Now grep-ing the rest of the handler layer for other `var x []T` return values that might be hiding the same bug.

## Resources That Helped

- [Go by Example: Slices](https://gobyexample.com/slices)
- Go's `encoding/json` docs on nil value encoding

---
*I'm learning Go in public. Full code: github.com/absep98/Go_learn*
*Building a personal analytics backend from scratch.*
