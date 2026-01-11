//go:build !integration
// +build !integration

package bitbucket

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitGetBitbucket_Init_Success(t *testing.T) {
	// Set up environment variables
	originalUsername := os.Getenv("BITBUCKET_USERNAME")
	originalToken := os.Getenv("BITBUCKET_TOKEN")
	defer func() {
		if originalUsername != "" {
			os.Setenv("BITBUCKET_USERNAME", originalUsername)
		} else {
			os.Unsetenv("BITBUCKET_USERNAME")
		}
		if originalToken != "" {
			os.Setenv("BITBUCKET_TOKEN", originalToken)
		} else {
			os.Unsetenv("BITBUCKET_TOKEN")
		}
	}()

	os.Setenv("BITBUCKET_USERNAME", "test-user")
	os.Setenv("BITBUCKET_TOKEN", "test-token-123")

	gitProvider := &GitGetBitbucket{}
	result := gitProvider.Init()

	assert.True(t, result)
	assert.Equal(t, "test-user", gitProvider.username)
	assert.Equal(t, "test-token-123", gitProvider.token)
}

func TestGitGetBitbucket_Init_MissingUsername(t *testing.T) {
	// This test would cause os.Exit(1), so we skip it in unit tests
	t.Skip("Skipping test that calls os.Exit(1)")
}

func TestGitGetBitbucket_Init_MissingToken(t *testing.T) {
	// This test would cause os.Exit(1), so we skip it in unit tests
	t.Skip("Skipping test that calls os.Exit(1)")
}

func TestGenerateProjectKey(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "myproject",
			expected: "MYPROJECT",
		},
		{
			name:     "name with spaces",
			input:    "my project",
			expected: "MYPROJECT",
		},
		{
			name:     "name with hyphens",
			input:    "my-project-name",
			expected: "MYPROJECTNAME",
		},
		{
			name:     "name with special chars",
			input:    "my@project#name!",
			expected: "MYPROJECTNAME",
		},
		{
			name:     "name with underscores",
			input:    "my_project_name",
			expected: "MY_PROJECT_NAME",
		},
		{
			name:     "mixed case with numbers",
			input:    "MyProject123",
			expected: "MYPROJECT123",
		},
		{
			name:     "name with dots",
			input:    "my.project.name",
			expected: "MYPROJECTNAME",
		},
		{
			name:     "complex name",
			input:    "My-Project_Name.123!@#",
			expected: "MYPROJECT_NAME123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special chars",
			input:    "@#$%^&*()",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateProjectKey(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGitGetBitbucket_Auth(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestGitGetBitbucket_RepositoryExists(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestRepositoryExists_PackageFunction(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestProjectExists(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestGitGetBitbucket_CreateRepository(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestCreateRepository_PackageFunction(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestGitGetBitbucket_FetchOwnerRepos(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}

func TestFetchOwnerRepos_PackageFunction(t *testing.T) {
	// This test requires actual Bitbucket API or mocking at HTTP level
	t.Skip("Requires Bitbucket API mocking or integration test")
}
