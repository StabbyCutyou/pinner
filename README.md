# Pinner [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/StabbyCutyou/pinner)
A proof of concept approach for pinning dependencies in libraries and services / applications
in go. It works by manipulating the files in $GOPATH/src to via any available semver tags.

# Overview
This is a very quick, not complete, full of holes proof of concept of code to used
for a system of pinning an applications dependencies to a specific version based
on git tags.

It has several places that are functionally incomplete, but the idea was to given
a quick demonstration of how I think it could start to be implemented.

Use at your own risk, but hey - please give it a shot and provide some feedback!

Since the overall approach here will make changes to the src directory in your $GOPATH,
Pinner is likely more suited for projects using an approach like [gb](getgb.io),
which manages a unique $GOPATH environment per project.

Pinner expects to find tags with the prefix 'v', followed by a valid Semantic Version number.
If it does not find these tags, it will not be able to determine which version to Use
when resolving the dependency chain.

# In depth
The idea here is that each library or service would define a "pin" package in it's
own executable that would interact with something like this (pinner). You would
register the dependencies, and the version + constraint (~>1.0, <3.0, etc).

The current implementation will:

* Resolve all dependencies, break on any errors
* For each of those dependencies, check for a "pin" package, resolve those
    * Continue doing this until you run out of "pin" packages, or until you hit a breaking issue
* Build a map of dependencies to constraints
* Now, retrieve all dependencies, and resolve to find the version-tag that best fulfils all the provided constraints
  * Return an errror if none can be resolved
* Copy the files from the staginig area into the real gopath

# Sample Integration

I've provided a what would look like a sample Registry in pinner_test.go, which relies
directly on 2 dummy libraries I've set up, which themselves have a set of dependencies on
3 other libraries. Checkout [LibA]](https://github.com/StabbyCutyou/lib_a) or [LibB](https://github.com/StabbyCutyou/lib_b) to see examples.

A copy of a pin/main.go is provided here, as well:

```golang
package main

import "github.com/StabbyCutyou/pinner"

func main() {
	pinner.Register("github.com/StabbyCutyou/lib_c", "~> 1.0")
	pinner.Register("github.com/StabbyCutyou/lib_d", "= 1.2")

	pinner.RunMain()
}
```

## A note about the sample libraries

The sample libraries themselves are basically strawmn, holding a place for real libraries.
They exists only to demonstrate how Pinner works, as as such have no actual code,
outside of the pin/main.go registry. Additionally, they have several tags, all of
which (partially due to laziness, partially because it was all that I needed at the time)
share the same commit. Pinner, when checking out tags, leaves the repo in a detached state.
One interesting side effect of git checkout with detached branches is that if you checkout
tag A, then checkout tag B, and A and B share the same HEAD, ```git branch``` will still
appear as though you're checked out of tag A.

In other words, depending on the library and it's tags, you may notice (as with the sample
integration I have provided here) that it looks like it isn't checked out to the right tag
correctly if you inspect the staging area manually, but as best I can tell this is simply
a quirk of using git with a detachd branch.

# The dreaded "Diamond" Dependency

Pinner makes it possible to have a dependency tree resolve down a single version which
satisfies all the declared constraints for a given library.

It does not make it possible to "solve" dependencies which are mutually exclusive
of one another, like a "Diamond" Dependency.

To give a more concrete example, imagine you pin a requirement on LibA and LibB

```golang
pinner.Register("github.com/StabbyCutyou/lib_a", "~> 1.0")
pinner.Register("github.com/StabbyCutyou/lib_b", "= 1.2")
```

Both LibA and LibB define a dependecy on libC

LibA:
```golang
pinner.Register("github.com/StabbyCutyou/lib_c", "~> 2.0")
```

LibB:
```golang
pinner.Register("github.com/StabbyCutyou/lib_c", "= 1.0")
```

Since it is impossible to have a version that can satisfy both = 1.0 and ~> 2.0,
you cannot resolve the dependencies successfully. The only course of action here is to
fork one of the libraries, and change them to support the other version of LibC. Ideally,
you'd find a way to contribute this back upstream, but thats up to you.

# Roadmap

Right now, this is experimental and should be used for gathering design feedback
and to test the viability of the idea overall. I would greatly appreciate any
feedback or comments, so feel free to open Github issues or PRs as a way to provide this.
