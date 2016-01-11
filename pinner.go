// Package pinner allows you to pin packages
package pinner

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	// ErrUnsupportedDependency is returned whenever a depdency cannot be interpreted
	// TODO this is too generic and lacks any kind of helpful context
	ErrUnsupportedDependency = errors.New("Unsupported Dependency")
)

var rootLibraryPath = os.Getenv("GOPATH") + "/" + "pin_staging/"
var lfByte byte = 10

// registry holds the actual registry of each dependency
var registry = make([]*registryEntry, 0)

type registryEntry struct {
	libName    string
	constraint *version.Constraint
}

func (r *registryEntry) fetchGit() error {
	cloneURL := "https://" + r.libName + ".git"
	libDir := rootLibraryPath + r.libName

	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		// Create the directory, then clone into it
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
	}

	cmd := exec.Command("git", "fetch")
	cmd.Dir = libDir

	_, err := cmd.Output()

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

	// Now we've got all the versions for the top level dependencies
	// Go through and checkout the right version
	for _, v := range versions {
		if r.constraint.Check(v) {
			cmd := exec.Command("git", "checkout", "v"+v.String())
			cmd.Dir = libDir

			_, err := cmd.Output()
			if err != nil {
				return err
			}
			break
		}
	}

	// Now, for each of those libraries, get their dependencies, and fetch them
	deps, err := r.getStagingLibraryDependencies()
	if err != nil {
		return err
	}

	for _, d := range deps {
		err = d.fetch()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *registryEntry) getStagingLibraryDependencies() ([]*registryEntry, error) {
	libDir := rootLibraryPath + r.libName
	results := make([]*registryEntry, 0)
	pinMain := libDir + "/pin/main.go"
	if _, err := os.Stat(pinMain); os.IsNotExist(err) {
		// no pin main, no way to tell deps. Log a message, return empty set
		return results, nil
	}
	log.Print("go run " + pinMain)
	cmd := exec.Command("go", "run", pinMain)
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	log.Print(string(output))
	cache := make([]byte, 0)
	for _, o := range output {
		if o == lfByte {
			// Save the cache value if and only if it's not "list"
			line := string(cache)

			parts := strings.SplitN(line, " ", 2)
			constraint, err := version.NewConstraint(parts[1])
			if err != nil {
				return nil, err
			}
			results = append(results, &registryEntry{
				libName:    parts[0],
				constraint: constraint[0],
			})

			// Now, reset the cache
			cache = make([]byte, 0)
		} else {
			cache = append(cache, o)
		}
	}

	return results, nil
}

// parseGitTag is used to help parse the output of git tag
// TODO find a way to refact this, as it's the same basic logic as in getStagingLibraryDependencies
func parseGitTag(output []byte) ([]string, error) {
	cache := make([]byte, 0)
	results := make([]string, 0)
	for _, o := range output {
		if o == lfByte {
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

// Report prints the registry to stdout
func Report() {
	for _, r := range registry {
		fmt.Print(fmt.Sprintf("%s %s\n", r.libName, r.constraint))
	}
}

func (r *registryEntry) fetch() error {
	if strings.Contains(r.libName, "github.com") {
		return r.fetchGit()
	}

	// The weakness here is that you need a handler per possible repository system
	// Bundler/Ruby has this easy - by now there is a ubiquitous and free public
	// system, as well as several private systems that are all integrated. For now
	// this remains an imperfect approach, but since a majority of all go libraries seem to be
	// hosted on github, I feel it's both a fair start, and indicative of the general
	// aproach you'd take for any other service

	return ErrUnsupportedDependency
}

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
	return nil
}

// RunMain is meant to be invoked from a pin/main.go binary
func RunMain() {
	mode := os.Getenv("GOPIN_MODE")
	if mode == "" || mode == "report" {
		Report()
		os.Exit(0)
	} else if mode == "pin" {
		errs := Pin()
		if errs != nil && len(errs) > 0 {
			for _, e := range errs {
				fmt.Print(e)
			}
			os.Exit(1)
		}
	}

	log.Fatal("Unsupported GOPIN_MODE: " + mode)
}
