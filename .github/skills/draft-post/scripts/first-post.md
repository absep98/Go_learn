---
title: Why I Decided to Learn Go — And What My First Commit Looked Like
subtitle: Part 1 of Learning Go in Public
tags: go, buildinpublic, backend, beginners
series: Learning Go in Public
---

On November 5, 2025, I pushed my first commit to a repo called `Go_learn`.

The commit message: *"Started learning Go lang scratch lessons."*

54 files. 5,365 lines. Variables, loops, structs, maps, goroutines, interfaces — everything I'd learned in a week, all dumped into one massive push because I didn't know any better.

That repo is still public. The messy files are still there. And this is the first post in a series where I'm going to walk through every stage of what happened next.

## Why Go?

I was coming from JavaScript and the MERN stack. I could build things. But I kept hitting a wall — I didn't truly understand what was happening under the hood. Express hides the HTTP details. Mongoose hides the database details. npm hides... everything.

I wanted a language that would force me to understand things, not abstract them away.

Go kept showing up. Backend roles I was interested in listed it. The standard library supposedly had everything you needed. No frameworks, no magic. And the syntax looked almost too simple to be real.

So I opened VS Code and typed:

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

And immediately got confused.

## The First Week: Where's My Semicolons?

Coming from JavaScript, Go felt alien. Some things that tripped me up in the first few days:

**Exported vs unexported names.** I spent an embarrassing amount of time wondering why `myVariable` in one file wasn't accessible from another. In Go, only names starting with a capital letter are exported. In JavaScript, everything is just... there.

**No classes.** Go has structs and methods, but no inheritance. I kept trying to think in OOP patterns and Go kept saying "no, just use composition." It took a while to stop fighting it.

**Error handling everywhere.** My first Go functions looked like this:

```go
result, err := someFunction()
if err != nil {
    fmt.Println("something went wrong:", err)
    return
}
```

I wrote `if err != nil` about 200 times that first week. It felt tedious. (Spoiler for a future post: I now think it's one of Go's best features.)

**No generics... then generics.** I learned Go right after generics were added, but every tutorial and StackOverflow answer was from the pre-generics era. Confusing when you're starting out.

## What My First Commit Actually Contained

I organized my learning into numbered folders:

```
01-basics/         → hello.go, types.go, functions.go
02-control-flow/   → loops.go, switch.go
03-data-structures → structs.go, maps.go, interfaces.go
05-concurrency/    → goroutines-basic.go, channels-basic.go, worker-pool.go
06-projects/       → number-guessing-game.go, todo-list.go
07-Pointers/       → pointers.go, nil-pointers-explained.go
08-errors/         → errors.go
09-packages/       → go modules guide
```

Yes, I skipped folder 04. No, I don't know why. It's still missing.

Every file is packed with comments — I was essentially writing tutorials for myself. Here's an actual snippet from `structs.go`:

```go
// Structs are like blueprints for creating custom types
// Think of them like objects in JavaScript, but typed and explicit
type Person struct {
    Name string
    Age  int
}
```

These files are messy, verbose, and sometimes wrong. I'm keeping every single one of them in the repo. They're the proof that I started from zero.

## The Moment It Clicked

By the end of week one, I built two small projects: a **number guessing game** and a **CLI todo list**. Neither is impressive. But writing them forced me to combine everything — structs, slices, loops, user input, error handling — into something that actually ran.

That's when Go stopped feeling like syntax to memorize and started feeling like a tool I could build with.

## What's Coming Next

In the next post, I'll talk about the decision that accelerated everything: **stopping the tutorials and building a real backend project.** I started a personal analytics REST API using only Go's standard library — no gin, no fiber, no frameworks. Just `net/http`, `database/sql`, and a lot of `if err != nil`.


Follow along — I'll be posting each stage of this journey as its own article. The messy commits, the breakthroughs, the mistakes. All of it.

*This is Part 1 of the "Learning Go in Public" series.*
