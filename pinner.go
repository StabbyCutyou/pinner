// Package pinner allows you to pin packages
package pinner

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
)

var rootLibraryPath = os.Getenv("GOPATH") + "/" + "test/"

type registryEntry struct {
	libName    string
	constraint *version.Constraint
}

func (r *registryEntry) fetchGit() error {
	cloneURL := "https://" + r.libName + ".git"
	libDir := rootLibraryPath + r.libName

	err := os.MkdirAll(libDir, 0700)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", cloneURL, ".")
	cmd.Dir = libDir
	_, err = cmd.Output()

	if err != nil {
		return err
	}

	cmd = exec.Command("git", "fetch")
	cmd.Dir = libDir

	_, err = cmd.Output()

	if err != nil {
		return err
	}

	cmd = exec.Command("git", "tag")
	cmd.Dir = libDir

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	tags, err := parseGitTag(output)

	// Sort the list of versions
	// Find the highest ranking version that fulfills the constraint
	// in a sorted list, this would be the first match

	versions := make([]*version.Version, len(tags))
	for i, t := range tags {
		ver := ""
		if t[0] == byte('v') {
			ver = t[1:]
		} else {
			ver = t
		}
		v, _ := version.NewVersion(ver)
		versions[i] = v
	}

	for _, v := range versions {
		if r.constraint.Check(v) {
			cmd := exec.Command("git", "checkout", "v"+v.String())
			cmd.Dir = libDir

			_, err := cmd.Output()
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

// parseGitTag is used to help parse the output of git tag
func parseGitTag(output []byte) ([]string, error) {
	cache := make([]byte, 0)
	results := make([]string, 0)
	for _, o := range output {
		if o == 10 { // 10 is the raw value
			// Save the cache value if and only if it's not "list"
			constraint := string(cache)
			// ignore tags that aren't versions
			if constraint[0] == 'v' {
				results = append(results, constraint)
			}

			// Now, reset the cache
			cache = make([]byte, 0)
		} else {
			cache = append(cache, o)

		}
	}

	return results, nil
}

func (r *registryEntry) fetch() error {
	if strings.Contains(r.libName, "github.com") {
		return r.fetchGit()

	} else {
		// The weakness here is that you need a handler per possible repository system
		// Bundler/Ruby has this easy - by now there is a ubiquitous and free public
		// system, as well as several private systems that are all integrated. For now
		// this remains an imperfect approach, but since a majority of all go libraries seem to be
		// hosted on github, I feel it's both a fair start, and indicative of the general
		// aproach you'd take for any other service
	}

	return nil
}

var registry = make([]*registryEntry, 0)

// Register stores information about versions to check, which will apply
// once you call Check()
func Register(name string, constraint string) error {

	lockVer, err := version.NewConstraint(constraint)
	if err != nil {
		log.Print(err)
		return err
	}
	reg := &registryEntry{
		constraint: lockVer[0],
		libName:    name,
	}
	registry = append(registry, reg)
	return nil
}

// Pin is meant to be called once you have finished Registering your versions.
// Typically, you would do your registering in a seprate binary, which is specific
// to a given library or application. This would naturally wreak havoc in default
// environments, where all projects share the same GOPATH, but would work in environments
// like gb, where each project can have it's own individual workspace. A possible
// solution to that would be to have each package keep a subdirectory of versions,
// so that each project could import the correct package. This would require either
// adding the version to the import, or some kind of deeper coupling of a dependency
// registry system with the actual underlying parts of go that import libraries.
// As such - it's not really in scope for what I'm doing here.
func Pin() []error {
	var errs []error
	for i := 0; i < len(registry); i++ {
		r := registry[i]
		err := r.fetch()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
