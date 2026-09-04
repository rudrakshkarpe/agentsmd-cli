package cli

import (
	"fmt"
	"os"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

func (a *app) requireProject() (*project.Project, error) {
	return project.Require(a.root)
}

func writeArtifact(p *project.Project, content string, force bool) error {
	if !force {
		if _, err := os.Stat(p.ArtifactPath()); err == nil {
			return fmt.Errorf("%s already exists; use --force to overwrite", project.Artifact)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return project.AtomicWrite(p.ArtifactPath(), []byte(content), 0o644)
}
