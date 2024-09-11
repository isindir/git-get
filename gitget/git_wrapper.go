package gitget

import (
	gitPlumbing "github.com/go-git/go-git/v5/plumbing"
)

type RepositoryI interface {
	Reference(name gitPlumbing.ReferenceName, resolved bool) (*gitPlumbing.Reference, error)
	Tag(name string) (*gitPlumbing.Reference, error)
}

type Repository struct{}
