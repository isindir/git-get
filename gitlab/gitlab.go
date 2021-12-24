/*
Copyright © 2021 Eriks Zelenka <isindir@users.sourceforge.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package gitlab

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"os"
	"strings"
)

func gitlabAuth(repositorySha string, baseUrl string) *gitlab.Client {
	token, tokenFound := os.LookupEnv("GITLAB_TOKEN")
	if !tokenFound {
		log.Fatalf("%s: Error - environment variable GITLAB_TOKEN not found", repositorySha)
		os.Exit(1)
	}

	clientOptions := gitlab.WithBaseURL("https://" + baseUrl)
	git, err := gitlab.NewClient(token, clientOptions)
	if err != nil {
		log.Fatalf("%s: Error - while trying to authenticate to Gitlab: %s", repositorySha, err)
		os.Exit(1)
	}

	return git
}

// ProjectExists checks if project exists and returns boolean if API call is successful
func ProjectExists(repositorySha string, baseUrl string, projectName string) bool {
	log.Debugf("%s: Checking repository '%s' '%s' existence", repositorySha, baseUrl, projectName)
	git := gitlabAuth(repositorySha, baseUrl)

	prj, _, err := git.Projects.GetProject(projectName, nil, nil)

	log.Debugf("%s: project: '%+v'", repositorySha, prj)

	if err != nil {
		return false
	} else {
		return true
	}
}

func GetProjectNamespace(repositorySha string, baseUrl string, projectNameFullPath string) (*gitlab.Namespace, string) {
	log.Debugf("%s: Getting Project FullPath Namespace '%s'", repositorySha, projectNameFullPath)
	git := gitlabAuth(repositorySha, baseUrl)

	pathElements := strings.Split(projectNameFullPath, "/")
	// Remove short project name from the path elements list
	if len(pathElements) > 0 {
		pathElements = pathElements[:len(pathElements)-1]
	}
	namespaceFullPath := strings.Join(pathElements, "/")

	namespaceObject, _, err := git.Namespaces.GetNamespace(namespaceFullPath, nil, nil)

	log.Debugf("%s: Getting namespace '%s': resulting namespace object: '%+v'", repositorySha, namespaceFullPath, namespaceObject)

	if err != nil {
		return nil, namespaceFullPath
	} else {
		return namespaceObject, namespaceFullPath
	}
}

// CreateProject - Create new code repository
func CreateProject(
	repositorySha string,
	baseUrl string,
	projectName string,
	namespaceID int,
	mirrorVisibilityMode string,
	sourceURL string,
) *gitlab.Project {
	git := gitlabAuth(repositorySha, baseUrl)

	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String(projectName),
		Description:          gitlab.String(fmt.Sprintf("Mirror of the '%s'", sourceURL)),
		MergeRequestsEnabled: gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.VisibilityValue(mirrorVisibilityMode)),
	}
	if namespaceID != 0 {
		p.NamespaceID = gitlab.Int(namespaceID)
	}

	project, _, err := git.Projects.CreateProject(p)
	if err != nil {
		log.Fatalf("%s: Error - while trying to create gitlab project '%s': '%s'", repositorySha, projectName, err)
		os.Exit(1)
	}

	return project
}
