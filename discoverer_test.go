package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetModules_ReturnsErrorFromListModules(t *testing.T) {
	mockExecutor := MockExecutor{RunError: fmt.Errorf("an-error-from-executor")}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.GetModules()
	require.Error(t, err)

	assert.Contains(t, err.Error(), "listing modules: ")
}

func Test_GetModules_ReturnsErrorFromParseModules(t *testing.T) {
	mockExecutor := MockExecutor{
		CommandOutput: "invalid-output",
	}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.GetModules()
	require.Error(t, err)

	assert.Contains(t, err.Error(), "parsing modules: ")
}

func Test_ListModules_CallsExecutorRun(t *testing.T) {
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.listModules()
	require.NoError(t, err)

	runCalls := mockExecutor.RunCalls
	require.Len(t, runCalls, 1)

	listArgs := "list -m -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}==START=={{.Path}},{{.Version}},{{.Update.Version}}==END=={{end}}' all"
	assert.Equal(t, runCalls[0], RunCall{
		Command: "go",
		Args:    listArgs,
	})
}

func Test_ListModules_ReturnsErrorFromExecutor(t *testing.T) {
	wantError := fmt.Errorf("an-error-from-executor")
	mockExecutor := MockExecutor{RunError: wantError}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.listModules()

	assert.Error(t, wantError, err)
}

func Test_ListModules_ReturnsModules(t *testing.T) {
	moduleName := "a-module-name"
	wantModules := []Module{
		{
			Name: moduleName,
		},
	}

	modulesListOutput := modulesToListFormat(wantModules...)
	mockExecutor := MockExecutor{
		CommandOutput: modulesListOutput,
	}

	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	moduleOutput, err := d.listModules()
	require.NoError(t, err)

	assert.Equal(t, modulesListOutput, moduleOutput)
}

func Test_ListModules_HandlesLatestModules(t *testing.T) {
	commandOutput := []string{
		"go: finding golang.org/x/sync latest",
		"go: finding golang.org/x/net latest",
		"go: finding gopkg.in/tomb.v1 latest",
	}

	result := strings.Join(commandOutput, "\n")
	mockExecutor := MockExecutor{
		CommandOutput: result,
	}

	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	moduleOutput, err := d.listModules()
	require.NoError(t, err)

	assert.Equal(t, result, moduleOutput)
}

func Test_ParseModules_ReturnsErrorWhenInvalidModuleRegex(t *testing.T) {
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{}, WithModuleRegex("not a valid regex ("))

	_, err := d.parseModules("")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error parsing regexp")
}

func Test_ParseModules_ReturnsErrorWhenNotAllMatched(t *testing.T) {
	output := "===START===example.com/a/module,1.0.0===END==="
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "regex was not able to find all matches")
}

func Test_ParseModules_ReturnsErrorWhenFromVersionIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:      "a-module-name",
		ToVersion: semver.MustParse("1.0.0"),
	}
	output := moduleToListFormat(wantModule)
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), fmt.Sprintf("parsing from version %q:", wantModule.FromVersion))
}

func Test_ParseModules_ReturnsErrorWhenToVersionIsNotAValidSemver(t *testing.T) {
	wantModule := Module{
		Name:        "a-module-name",
		FromVersion: semver.MustParse("1.0.0"),
	}
	output := moduleToListFormat(wantModule)
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

	_, err := d.parseModules(output)
	require.Error(t, err)

	assert.Contains(t, err.Error(), fmt.Sprintf("parsing to version %q:", wantModule.ToVersion))
}

func Test_ParseModules_ReturnsExpectedModules(t *testing.T) {
	wantModules := []Module{
		{
			Name:         "a-minor-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.1.0"),
			MinorUpgrade: true,
		},
		{
			Name:         "a-minor-upgrade-with-patch-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.1.1"),
			PatchUpgrade: false,
			MinorUpgrade: true,
		},
		{
			Name:         "a-patch-upgrade",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.0.1"),
			PatchUpgrade: true,
			MinorUpgrade: false,
		},
	}
	mockExecutor := MockExecutor{}
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})
	moduleListOutput := modulesToListFormat(wantModules...)
	modules, err := d.parseModules(moduleListOutput)
	require.NoError(t, err)
	require.Len(t, modules, 3)

	assert.Equal(t, wantModules, modules)
}

func Test_ParseModules_SkipsEmptyModuleLines(t *testing.T) {
	wantModules := []Module{
		{
			Name:         "a-module-name",
			FromVersion:  semver.MustParse("1.0.0"),
			ToVersion:    semver.MustParse("1.1.0"),
			MinorUpgrade: true,
		},
		{
			Name:        "another-module-name",
			FromVersion: semver.MustParse("1.0.0"),
			ToVersion:   semver.MustParse("3.0.0"),
		},
	}
	var mockExecutor MockExecutor
	d := NewDiscoverer(&mockExecutor, &MockHTTPClient{})

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

func Test_GetChangelog__CallsGivenHttpClient(t *testing.T) {
	name := "github.com/stretchr/testify"
	given := Module{
		Name: name,
	}

	mockClient := NewMockHTTPClient()
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, _ = d.GetChangelog(given)

	calls := mockClient.GetCalls()
	assert.Len(t, calls, 1)
}

func Test_GetGithubRepoFromModule_ReturnsExpectedModule(t *testing.T) {
	wantRepo := "a-project/a-wantRepo-name"
	m := Module{
		Name: fmt.Sprintf("github.com/%s", wantRepo),
	}
	gotRepo, err := getGithubRepoFromModule(m)
	require.NoError(t, err)

	assert.Equal(t, wantRepo, gotRepo)
}

func Test_GetChangelog__CallsHttpClientWithExpectedQueryParams(t *testing.T) {
	repo := "stretchr/testify"
	name := fmt.Sprintf("github.com/%s", repo)
	given := Module{
		Name:         name,
		PatchUpgrade: false,
		MinorUpgrade: false,
	}

	mockClient := NewMockHTTPClient()
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, _ = d.GetChangelog(given)

	calls := mockClient.GetCalls()
	require.Len(t, calls, 1)

	url := calls[0].URL
	require.NotNil(t, url)
	queryParams := url.RawQuery

	wantSearch := fmt.Sprintf("repo:%s+filename:CHANGELOG.md", repo)
	assert.Contains(t, queryParams, wantSearch)
}

func Test_GetChangelog_ReturnsErrorFromClient(t *testing.T) {
	given := Module{
		Name:         "github.com/stretchr/testify",
		PatchUpgrade: false,
		MinorUpgrade: false,
	}

	mockClient := NewMockHTTPClient()
	wantError := fmt.Errorf("an error from the HTTP Client")
	mockClient.GivenErrorIsReturned(wantError)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, err := d.GetChangelog(given)
	require.Error(t, err)

	assert.EqualError(t, err, fmt.Sprintf("failed to make a request for changelog: %s", wantError))
}

func Test_GetChangelog_ReturnsMatchingErrorWhenCannotParseModuleName(t *testing.T) {
	d := NewDiscoverer(&MockExecutor{}, &MockHTTPClient{})

	_, err := d.GetChangelog(Module{Name: "not-a-valid-module-name"})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unable to parse module name")
}
func Test_GetChangelog_ReturnsUnmarshallingErrorWhenResponseInvalid(t *testing.T) {
	mockClient := NewMockHTTPClient()
	mockClient.GivenResponseIsReturned(200, "not-valid-json", nil)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, err := d.GetChangelog(Module{
		Name: "github.com/foo/bar",
	})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unexpected response from github API:")
}

func Test_GetChangelog_ReturnsExpectedURL(t *testing.T) {
	module := newValidModule()
	wantURL := "url"
	githubResponse := GithubFileSearchResponse{
		TotalCount: 1,
		Items: []Item{
			{Name: module.Name, Path: "CHANGELOG.md", HTMLURL: wantURL},
		},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := NewMockHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	gotChangelog, err := d.GetChangelog(module)
	require.NoError(t, err)

	assert.Equal(t, wantURL, gotChangelog)
}

func newValidModule() Module {
	return Module{
		Name: "github.com/project/repo",
	}
}

func Test_GetChangelog_ReturnsRootChangelogIfMultipleFound(t *testing.T) {
	githubResponse := GithubFileSearchResponse{
		TotalCount: 2,
		Items: []Item{
			{Name: "another module", Path: "another-path", HTMLURL: "two-url"},
			{Name: "name", Path: "CHANGELOG.md", HTMLURL: "one-url"},
		},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := NewMockHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	changelog, err := d.GetChangelog(newValidModule())
	require.NoError(t, err)

	assert.Equal(t, githubResponse.Items[1].HTMLURL, changelog)
}

func Test_GetChangelog_ReturnsErrorWhenChangelogIsNotFound(t *testing.T) {
	githubResponse := GithubFileSearchResponse{
		TotalCount: 1,
		Items: []Item{
			{Name: "name", Path: "a-path", HTMLURL: "one-url"},
		},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := NewMockHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, err = d.GetChangelog(newValidModule())
	require.Error(t, err)

	assert.Contains(t, err.Error(), "failed to find a root level CHANGELOG.md")
}

func Test_GetChangelog_ReturnsErrorWhenNoSearchResultsFound(t *testing.T) {
	githubResponse := GithubFileSearchResponse{
		TotalCount: 0,
		Items:      []Item{},
	}
	body, err := json.Marshal(githubResponse)
	require.NoError(t, err)

	mockClient := NewMockHTTPClient()
	mockClient.GivenResponseIsReturned(200, string(body), nil)
	d := NewDiscoverer(&MockExecutor{}, mockClient)

	_, err = d.GetChangelog(newValidModule())
	require.Error(t, err)

	assert.Contains(t, err.Error(), "failed to find a root level CHANGELOG.md")
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
