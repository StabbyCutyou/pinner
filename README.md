# Pinner
A proof of concept approach for pinning dependencies in libraries and services / applications
in go. It works by manipulating the files in $GOPATH/src to via any available semver tags.

# Overview
This is a very quick, not complete, full of holes proof of concept of code to used
for a system of pinning an applications dependencies to a specific version based
on git tags.

It has several places that are functionally incomplete, but the idea was to given
a quick demonstration of how I think it could start to be implemented.

# In depth
The idea here is that each library or service would define a "pin" package in it's
own executable that would interact with something like this (pinner). You would
register the dependencies, and the version + constraint (~>1.0, <3.0, etc).

The current implementation simply goes and fetches them, but a proper implementation
would need to do the following:

* Resolve all dependencies, break on any errors
* For each of those dependencies, check for a "pin" package, resolve those
    * Continue doing this until you run out of "pin" packages, or until you hit a breaking issue
* Build a tree of dependencies, such that you know both what each library depends on all the way down
* From this tree, coalesce a list of unique dependencies with their version constraints
* Now, feed this list to the current process, where it will actually retrieve all dependencies, and resolve to find the version-tag that best fulfils that constraint.
