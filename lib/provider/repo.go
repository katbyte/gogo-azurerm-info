package provider

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
)

type Repo struct {
	Path string
	Git  *git.Repository
}

func NewRepo(path string) (*Repo, error) {
	r := Repo{
		Path: path,
	}

	// open repo
	git, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("opening repo %s: %w", path, err)
	}
	r.Git = git

	return &r, nil
}

func (r Repo) CheckoutTag(tag string) error {
	t, err := r.Git.Tag(tag)
	if err != nil {
		return fmt.Errorf("getting tag %s: %w", tag, err)
	}

	wt, err := r.Git.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree: %w", err)
	}

	b := plumbing.ReferenceName(fmt.Sprintf("refs/heads/v/%s", tag))
	if exists, err := r.Git.Storer.Reference(b); err == nil {
		if exists != nil {
			// Delete the branch
			if err := r.Git.Storer.RemoveReference(b); err != nil {
				return fmt.Errorf("deleting branch %s: %w", b.String(), err)
			}
		}
	}

	// get correct hash
	h, err := r.Git.ResolveRevision(plumbing.Revision(t.Hash().String()))
	if err != nil {
		return fmt.Errorf("getting hash: %w", err)
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Hash:   *h,
		Force:  true,
		Create: true,
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/v/%s", tag)),
	})
	if err != nil {
		return fmt.Errorf("checking out %s: %w", tag, err)
	}

	return nil
}

func (r Repo) GetVersions() (*[]Version, error) {
	tags, err := r.Git.Tags()
	if err != nil {
		return nil, fmt.Errorf("getting tags for %s: %w", r.Path, err)
	}
	defer tags.Close()

	versionRegex := regexp.MustCompile(`v[0-9]+\.[0-9]+\.[0-9]+`)

	versionTags := []string{}
	tagCount := 0
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		v := ref.Name().Short()

		if versionRegex.MatchString(v) {
			versionTags = append(versionTags, v)
		}

		tagCount++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("iterating tags: %w", err)
	}

	// sort
	sort.Slice(versionTags, func(i, j int) bool {
		vi, _ := version.NewVersion(versionTags[i])
		vj, _ := version.NewVersion(versionTags[j])
		return vi.GreaterThan(vj)
	})

	versions := []Version{}
	for _, v := range versionTags {
		versions = append(versions, Version{
			Name: v,
			Path: r.Path,
		})
	}

	return &versions, nil
}
