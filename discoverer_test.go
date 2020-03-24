package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/frasercobb/gomo/mock"
)

func Test_ListModulesCallsExecutorRun(t *testing.T) {
	mockExecutor := &mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(mockExecutor),
	)

	_, err := d.listModules()
	require.NoError(t, err)

	runCalls := mockExecutor.RunCalls
	require.Len(t, runCalls, 1)

	listArgs := "list -m -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}==START=={{.Path}},{{.Version}},{{.Update.Version}}==END=={{end}}' all"
	assert.Equal(t, runCalls[0], mock.RunCall{
		Command: "go",
		Args:    listArgs,
	})
}

func Test_ListModulesReturnsErrorFromExecutor(t *testing.T) {
	wantError := fmt.Errorf("an-error-from-executor")
	mockExecutor := mock.Executor{RunError: wantError}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	_, err := d.listModules()

	assert.Error(t, wantError, err)
}

func Test_ListModulesReturnsModules(t *testing.T) {
	moduleName := "a-module-name"
	wantModules := []Module{
		{
			Name: moduleName,
		},
	}

	modulesListOutput := modulesToListFormat(wantModules...)
	mockExecutor := mock.Executor{
		CommandOutput: modulesListOutput,
	}

	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	moduleOutput, err := d.listModules()
	require.NoError(t, err)

	assert.Equal(t, modulesListOutput, moduleOutput)
}

func Test_ListModulesHandlesLatestModules(t *testing.T) {
	commandOutput := []string{
		"go: finding golang.org/x/sync latest",
		"go: finding golang.org/x/net latest",
		"go: finding gopkg.in/tomb.v1 latest",
	}

	result := strings.Join(commandOutput, "\n")
	mockExecutor := mock.Executor{
		CommandOutput: result,
	}

	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	moduleOutput, err := d.listModules()
	require.NoError(t, err)

	assert.Equal(t, result, moduleOutput)
}

func Test_ParseModulesReturnsErrorWhenInvalidModuleRegex(t *testing.T) {
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)
	d.ModuleRegex = "not a valid regex ("

	_, err := d.parseModules("")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error parsing regexp")
}

func Test_ParseModulesReturnsErrorWhenNotAllMatched(t *testing.T) {
	output := "===START===example.com/a/module,1.0.0===END==="
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "regex was not able to find all matches")
}

func Test_ParseModulesReturnsErrorWhenFromVersionIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:      "a-module-name",
		ToVersion: semver.MustParse("1.0.0"),
	}
	output := moduleToListFormat(wantModule)
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), fmt.Sprintf("parsing from version %q:", wantModule.FromVersion))
}

func Test_ParseModulesReturnsErrorWhenToVersionIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:        "a-module-name",
		FromVersion: semver.MustParse("1.0.0"),
	}
	output := moduleToListFormat(wantModule)
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), fmt.Sprintf("parsing to version %q:", wantModule.ToVersion))
}

func Test_ParseModulesReturnsExpectedModules(t *testing.T) {
	wantModules := []Module{
		{
			Name:         "a-major-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("2.0.0"),
			MajorUpgrade: true,
		},
		{
			Name:         "a-minor-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.1.0"),
			MinorUpgrade: true,
		},
		{
			Name:         "a-patch-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.0.1"),
			MajorUpgrade: false,
			MinorUpgrade: false,
		},
	}
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)
	moduleListOutput := modulesToListFormat(wantModules...)
	modules, err := d.parseModules(moduleListOutput)
	require.NoError(t, err)
	require.Len(t, modules, 3)

	assert.Equal(t, wantModules, modules)
}

func Test_ParseModulesSkipsEmptyModuleLines(t *testing.T) {
	wantModules := []Module{
		{
			Name:         "a-module-name",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.1.0"),
			MinorUpgrade: true,
		},
		{
			Name:         "another-module-name",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("3.0.0"),
			MajorUpgrade: true,
		},
	}
	mockExecutor := mock.Executor{}
	d := NewDiscoverer(
		WithExecutor(&mockExecutor),
	)

	var moduleListWithEmptyLines []string
	moduleListWithEmptyLines = append(moduleListWithEmptyLines, "")
	for _, module := range wantModules {
		moduleListWithEmptyLines = append(moduleListWithEmptyLines, moduleToListFormat(module))
	}
	moduleListWithEmptyLines = append(moduleListWithEmptyLines, "''")

	modules, err := d.parseModules(strings.Join(moduleListWithEmptyLines, "\n"))
	require.NoError(t, err)

	assert.Equal(t, wantModules, modules)
}

func Test_GetChangelogCallsGivenHttpClient(t *testing.T) {
	name := "github.com/stretchr/testify"
	given := Module{
		Name:         name,
		MajorUpgrade: false,
		MinorUpgrade: false,
	}

	mockClient := mock.NewHTTPClient()
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, _ = d.GetChangelog(given)

	calls := mockClient.GetCalls()
	assert.Len(t, calls, 1)
}

func Test_GetGithubRepoFromModule(t *testing.T) {
	wantRepo := "a-project/a-wantRepo-name"
	m := Module{
		Name: fmt.Sprintf("github.com/%s", wantRepo),
	}
	gotRepo, err := getGithubRepoFromModule(m)
	require.NoError(t, err)

	assert.Equal(t, wantRepo, gotRepo)
}

func Test_GetChangelogCallsHttpClientWithExpectedQueryParams(t *testing.T) {
	repo := "stretchr/testify"
	name := fmt.Sprintf("github.com/%s", repo)
	given := Module{
		Name:         name,
		MajorUpgrade: false,
		MinorUpgrade: false,
	}

	mockClient := mock.NewHTTPClient()
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, _ = d.GetChangelog(given)

	calls := mockClient.GetCalls()
	require.Len(t, calls, 1)

	url := calls[0].URL
	require.NotNil(t, url)
	queryParams := url.RawQuery

	wantSearch := fmt.Sprintf("repo:%s+filename:CHANGELOG.md", repo)
	assert.Contains(t, queryParams, wantSearch)
}

func Test_GetChangelogReturnsErrorFromClient(t *testing.T) {
	given := Module{
		Name:         "github.com/stretchr/testify",
		MajorUpgrade: false,
		MinorUpgrade: false,
	}

	mockClient := mock.NewHTTPClient()
	wantError := fmt.Errorf("an error from the HTTP Client")
	mockClient.GivenErrorIsReturned(wantError)
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, err := d.GetChangelog(given)
	require.Error(t, err)

	assert.EqualError(t, err, fmt.Sprintf("failed to make a request for changelog: %s", wantError))
}

func Test_GetChangelogReturnsUnmarshallingErrorWhenResponseInvalid(t *testing.T) {
	mockClient := mock.NewHTTPClient()
	mockClient.GivenResponseIsReturned(200, "not-valid-json", nil)
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, err := d.GetChangelog(Module{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unexpected response from github API:")
}

func Test_GetChangelogReturnsExpectedURL(t *testing.T) {
	module := Module{
		Name: "a-name",
	}
	wantURL := "url"
	githubResponse := GithubFileSearchResponse{
		TotalCount: 1,
		Items: []Item{
			{Name: module.Name, Path: "a-path", HTMLURL: wantURL},
		},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := mock.NewHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	gotChangelog, err := d.GetChangelog(module)
	require.NoError(t, err)

	assert.Equal(t, wantURL, gotChangelog)
}

func Test_GetChangelogReturnsErrorWhenMultipleSearchResultsFound(t *testing.T) {
	githubResponse := GithubFileSearchResponse{
		TotalCount: 2,
		Items: []Item{
			{Name: "name", Path: "a-path", HTMLURL: "one-url"},
			{Name: "another module", Path: "another-path", HTMLURL: "two-url"},
		},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := mock.NewHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, err = d.GetChangelog(Module{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "found more than one file search result:")
	assert.Contains(t, err.Error(), githubResponse.Items[0].HTMLURL)
	assert.Contains(t, err.Error(), githubResponse.Items[1].HTMLURL)
}

func Test_GetChangelogReturnsErrorWhenNoSearchResultsFound(t *testing.T) {
	githubResponse := GithubFileSearchResponse{
		TotalCount: 0,
		Items:      []Item{},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := mock.NewHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(
		WithHTTPClient(mockClient),
	)

	_, err = d.GetChangelog(Module{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "failed to find changelog")
}

func modulesToListFormat(modules ...Module) string {
	var result []string
	for _, module := range modules {
		result = append(result, moduleToListFormat(module))
	}
	return strings.Join(result, "\n")
}

func moduleToListFormat(module Module) string {
	return fmt.Sprintf("==START==%s,%s,%s==END==", module.Name, module.FromVersion, module.ToVersion)
}
