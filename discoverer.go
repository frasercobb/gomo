package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Module struct {
	Name         string
	FromVersion  *semver.Version
	ToVersion    *semver.Version
	MajorUpgrade bool
	MinorUpgrade bool
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Discoverer struct {
	Executor        Executor
	HTTPClient      HTTPClient
	ModuleRegex     string
	ListCommand     string
	ListCommandArgs []string
}

const (
	template           = "'{{if (and (not (or .Main .Indirect)) .Update)}}==START=={{.Path}},{{.Version}},{{.Update.Version}}==END=={{end}}'"
	expectedNumMatches = 4
)

type Option func(*Discoverer)

func NewDiscoverer(options ...Option) *Discoverer {
	d := &Discoverer{
		Executor:    &CommandExecutor{},
		ModuleRegex: "==START==(.+),(.+),(.+)==END==",
		ListCommand: "go",
		ListCommandArgs: []string{
			"list", "-m", "-u", "-f", template, "all",
		},
	}

	for _, option := range options {
		option(d)
	}

	return d
}

func WithExecutor(executor Executor) Option {
	return func(d *Discoverer) {
		d.Executor = executor
	}
}

func WithHTTPClient(client HTTPClient) Option {
	return func(d *Discoverer) {
		d.HTTPClient = client
	}
}

func (d *Discoverer) GetModules() ([]Module, error) {
	listOutput, err := d.listModules()
	if err != nil {
		return nil, fmt.Errorf("listing modules: %w", err)
	}

	modules, err := d.parseModules(listOutput)
	if err != nil {
		return nil, fmt.Errorf("parsing modules: %w", err)
	}

	return modules, nil
}

type GithubFileSearchResponse struct {
	TotalCount int    `json:"total_count"`
	Items      []Item `json:"items"`
}

type Item struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	HTMLURL string `json:"html_url"`
}

func (d *Discoverer) GetChangelog(module Module) (string, error) {
	u := &url.URL{
		Scheme:   "https",
		Host:     "api.github.com",
		Path:     "/search/code",
		RawQuery: fmt.Sprintf("q=repo:%s%sfilename:CHANGELOG.md", module.Name, "+"),
	}
	res, err := d.HTTPClient.Do(&http.Request{
		URL: u,
	})
	if err != nil {
		return "", fmt.Errorf("failed to make a request for changelog: %w", err)
	}
	var githubResp GithubFileSearchResponse
	decoder := json.NewDecoder(res.Body)
	if err = decoder.Decode(&githubResp); err != nil {
		return "", fmt.Errorf("unexpected response from github API: %w", err)
	}

	if len(githubResp.Items) > 1 {
		files := make([]string, len(githubResp.Items))
		for _, item := range githubResp.Items {
			files = append(files, item.HTMLURL)
		}
		return "", fmt.Errorf("found more than one file search result: %s", files)
	}

	return githubResp.Items[0].HTMLURL, nil
}

func (d *Discoverer) listModules() (string, error) {
	output, err := d.Executor.Run(d.ListCommand, d.ListCommandArgs...)
	if err != nil {
		return "", fmt.Errorf("running '%s %s': %w", d.ListCommand, d.ListCommandArgs, err)
	}
	return output, nil
}

func (d *Discoverer) parseModules(listOutput string) ([]Module, error) {
	re, err := regexp.Compile(d.ModuleRegex)
	if err != nil {
		return nil, err
	}

	var modules []Module
	modulesLines := strings.Split(listOutput, "\n")
	for _, line := range modulesLines {
		if isInvalidModuleLine(line) {
			continue
		}

		m, err := extractModule(line, re)
		if err != nil {
			return nil, err
		}

		modules = append(modules, m)
	}

	return modules, nil
}

func isInvalidModuleLine(line string) bool {
	if line == "''" {
		return true
	}
	if line == "" {
		return true
	}
	return false
}

func extractModule(moduleLine string, regex *regexp.Regexp) (Module, error) {
	matches := regex.FindStringSubmatch(moduleLine)
	if len(matches) != expectedNumMatches {
		return Module{}, fmt.Errorf("regex was not able to find all matches")
	}

	from, err := semver.NewVersion(matches[2])
	if err != nil {
		return Module{}, fmt.Errorf("parsing from version %q: %w", from, err)
	}

	to, err := semver.NewVersion(matches[3])
	if err != nil {
		return Module{}, fmt.Errorf("parsing to version %q: %w", to, err)
	}

	return Module{
		Name:         matches[1],
		FromVersion:  from,
		ToVersion:    to,
		MajorUpgrade: to.Major() > from.Major(),
		MinorUpgrade: to.Minor() > from.Minor(),
	}, nil
}
