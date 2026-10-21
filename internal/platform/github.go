package platform

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"errors"

	"github.com/google/go-github/v80/github"
	"golang.org/x/sync/errgroup"
)

const (
	githubRepoStars      = "stars"
	githubRepoWatchers   = "watchers"
	githubRepoForks      = "forks"
	githubRepoOpenIssues = "open_issues"

	githubScopeOwner = "owner"
	githubScopeAll   = "all"

	githubMaxPerPage = 100
)

// Github structure connects to the Github API.
type Github struct {
	client   *github.Client
	repoName string
	owner    string
}

// GithubClient to fetch Github related data.
func NewGithubClient(token string, owner string, repoName string) (*Github, error) {
	client := github.NewClient(nil).WithAuthToken(token)

	return &Github{
		client:   client,
		repoName: repoName,
		owner:    owner,
	}, nil
}

// TotalStars of a repository.
func (g *Github) TotalStars(repository string) (int, error) {
	r, err := g.fetchRepo(repository)
	if err != nil {
		return 0, err
	}

	return r.GetStargazersCount(), nil
}

// TotalWatchers of a repository overtime.
func (g *Github) TotalWatchers(repository string) (int, error) {
	r, err := g.fetchRepo(repository)
	if err != nil {
		return 0, err
	}

	return r.GetSubscribersCount(), nil
}

// TotalOpenIssues of a repository overtime.
func (g *Github) TotalOpenIssues(repository string) (int, error) {
	r, err := g.fetchRepo(repository)
	if err != nil {
		return 0, err
	}

	return r.GetOpenIssuesCount(), nil
}

// ListBranches of a repository.
func (g *Github) ListBranches(repository string, limit int) ([][]string, error) {
	headers := []string{"name"}

	bs, err := g.fetchBranches(repository, limit)
	if err != nil {
		return nil, err
	}

	if limit > len(bs) {
		limit = len(bs)
	}

	branches := make([][]string, limit+1)
	branches[0] = headers
	for k, v := range bs {
		n := "unknown"
		if v.Name != nil {
			n = *v.Name
		}

		if k < limit {
			branches[k+1] = append(branches[k+1], n)
		}
	}

	return branches, nil
}

// ListRepo of a Github account.
func (g *Github) ListRepo(limit int, order string, metrics []string) ([][]string, error) {
	headers := []string{"name"}

	rs, err := g.fetchAllRepo(order)
	if err != nil {
		return nil, err
	}

	if limit > len(rs) {
		limit = len(rs)
	}

	repos := make([][]string, limit+1)

	stars := false
	watchers := false
	forks := false
	openIssues := false

	if contains(metrics, githubRepoStars) {
		headers = append(headers, githubRepoStars)
		stars = true
	}
	if contains(metrics, githubRepoWatchers) {
		headers = append(headers, githubRepoWatchers)
		watchers = true
	}
	if contains(metrics, githubRepoForks) {
		headers = append(headers, githubRepoForks)
		forks = true
	}
	if contains(metrics, githubRepoOpenIssues) {
		headers = append(headers, githubRepoOpenIssues)
		openIssues = true
	}

	repos[0] = headers

	for k, v := range rs {
		if k < limit {
			repos[k+1] = append(repos[k+1], v.GetName())
			if stars {
				repos[k+1] = append(repos[k+1], strconv.FormatInt(int64(v.GetStargazersCount()), 10))
			}
			if watchers {
				repos[k+1] = append(repos[k+1], strconv.FormatInt(int64(v.GetSubscribersCount()), 10))
			}
			if forks {
				repos[k+1] = append(repos[k+1], strconv.FormatInt(int64(v.GetForksCount()), 10))
			}
			if openIssues {
				repos[k+1] = append(repos[k+1], strconv.FormatInt(int64(v.GetOpenIssuesCount()), 10))
			}
		}
	}

	return repos, nil
}

// ListIssues of a repository.
func (g *Github) ListIssues(repository string, limit int) ([][]string, error) {
	headers := []string{"name", "state"}

	is, err := g.fetchIssues(repository, limit)
	if err != nil {
		return nil, err
	}

	if limit > len(is) {
		limit = len(is)
	}

	issues := make([][]string, limit+1)
	issues[0] = headers
	for k, v := range is {
		n := "unknown"
		if v.Title != nil {
			n = *v.Title
		}

		state := "unknown"
		if v.State != nil {
			state = *v.State
		}

		if k < limit {
			issues[k+1] = append(issues[k+1], n)
			issues[k+1] = append(issues[k+1], state)
		}
	}

	return issues, nil
}

// ListPullRequests of a repository.
func (g *Github) ListPullRequests(repository string, limit int) ([][]string, error) {
	is, err := g.fetchPullRequests(repository, limit)
	if err != nil {
		return nil, err
	}

	lpr := formatListPullRequests(is, limit)

	return lpr, nil
}

func formatListPullRequests(is []*github.PullRequest, limit int) [][]string {
	if limit > len(is) {
		limit = len(is)
	}

	prs := [][]string{{"title", "state", "author", "created at"}}
	for _, v := range is[:limit] {
		state := v.GetState()
		if v.MergedAt != nil {
			state = "merged"
		}

		createdAt := "unknown"
		if v.CreatedAt != nil {
			createdAt = v.CreatedAt.Format("2006-01-02")
		}

		prs = append(prs, []string{v.GetTitle(), state, v.GetUser().GetLogin(), createdAt})
	}

	return prs
}

// Views on a github repository the last 7 days.
func (g *Github) Views(repository string, days int) ([]string, []int, error) {
	tv, err := g.fetchViews(repository)
	if err != nil {
		return nil, nil, err
	}

	counts := []int{}
	for _, v := range tv.Views {
		counts = append(counts, v.GetCount())
	}

	dimension := []string{}
	for _, v := range tv.Views {
		dimension = append(dimension, v.GetTimestamp().Format("01-02"))
	}

	return dimension, counts, nil
}

// CountCommits of a repository overtime.
func (g *Github) CountCommits(
	repository string,
	scope string,
	startWeek int64,
	endWeek int64,
	startDate time.Time,
) ([]string, []int, error) {
	c, err := g.fetchCommitCount(repository)
	if err != nil {
		return nil, nil, err
	}

	cm := c.Owner
	if scope == githubScopeAll {
		cm = c.All
	}

	d, co := formatCountCommits(cm, startWeek, endWeek, startDate)

	return d, co, nil
}

func formatCountCommits(
	c []int,
	startWeek int64,
	endWeek int64,
	startDate time.Time,
) ([]string, []int) {
	// Reverse the count of commits (from ASC to DESC).
	for i := len(c)/2 - 1; i >= 0; i-- {
		opp := len(c) - 1 - i
		c[i], c[opp] = c[opp], c[i]
	}

	counts := append([]int{}, c...)

	dimension := []string{}
	for k := range c {
		startWeekDay := int(startDate.Weekday())
		beginningOfWeek := -(startWeekDay)
		weekBefore := (7 * k)
		s := startDate.AddDate(0, 0, (beginningOfWeek - weekBefore))
		if startWeekDay == int(time.Sunday) && k == 0 {
			s = startDate
		}
		dimension = append(dimension, s.Format("01-02"))
	}

	d := dimension[endWeek:startWeek]
	co := counts[endWeek:startWeek]

	// reverse the count of commits (from DESC to ASC)
	for i := len(co)/2 - 1; i >= 0; i-- {
		opp := len(co) - 1 - i
		co[i], co[opp] = co[opp], co[i]
	}

	// reverse the dimensions (from DESC to ASC)
	for i := len(d)/2 - 1; i >= 0; i-- {
		opp := len(d) - 1 - i
		d[i], d[opp] = d[opp], d[i]
	}
	return d, co
}

// CountStars of a repository overtime.
// Only on a daily basis for now.
func (g *Github) CountStars(repository string, startDate, endDate time.Time) (dim []string, val []int, err error) {
	se, err := g.fetchStars(repository)
	if err != nil {
		return nil, nil, err
	}

	dim, val = formatCountStars(se, "01-02", startDate, endDate, false)
	return dim, val, nil
}

func formatCountStars(stargazers []*github.Stargazer, timeLayout string, startDate, endDate time.Time, addMissingDays bool) (dim []string, val []int) {
	sort.SliceStable(stargazers, func(i, j int) bool {
		return stargazers[i].StarredAt.Time.Before(stargazers[j].StarredAt.Time)
	})
	d, va := aggregateStarResults(stargazers)
	if addMissingDays {
		d, va = fillMissingDays(d, va)
	}

	for k, v := range d {
		if v.Before(endDate) && v.After(startDate) {
			dim = append(dim, v.Format(timeLayout))
			val = append(val, va[k])
		}
	}

	return dim, val
}

func aggregateStarResults(stargazers []*github.Stargazer) (dim []time.Time, val []int) {
	var countStar int
	for k, v := range stargazers {
		countStar++

		// If last element of the slice processed
		if len(stargazers) == k+1 {
			dim = append(dim, v.StarredAt.Time)
			val = append(val, countStar)
			return
		}

		processingDate := v.StarredAt.Time.Format("2006-01-02")
		nextDate := stargazers[k+1].StarredAt.Time.Format("2006-01-02")

		if processingDate != nextDate {
			dim = append(dim, v.StarredAt.Time)
			val = append(val, countStar)
			countStar = 0
		}
	}

	return
}

// fetchStars from the Github API. Every stars are fetched.
// Unfortunatelly, perPage is limited to 100, so we need multiple requests.
func (g *Github) fetchStars(repository string) (s []*github.Stargazer, err error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	r, err := g.fetchRepo(repository)
	if err != nil {
		return nil, err
	}

	pages := (*r.StargazersCount / githubMaxPerPage)
	if *r.StargazersCount%100 != 0 {
		pages += 1
	}

	var lock sync.Mutex
	var eg errgroup.Group
	sem := make(chan bool, 4)

	// At most 4 pages are fetched at the same time.
	for i := 1; i <= pages; i++ {
		sem <- true
		page := i
		eg.Go(func() error {
			defer func() { <-sem }()
			e, _, err := g.client.Activity.ListStargazers(context.Background(), g.owner, repo, &github.ListOptions{
				Page:    page,
				PerPage: githubMaxPerPage,
			})
			if err != nil {
				return fmt.Errorf("can't find repo %s of owner %s: %w", repo, g.owner, err)
			}

			lock.Lock()
			defer lock.Unlock()
			s = append(s, e...)
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (g *Github) fetchRepo(repository string) (*github.Repository, error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	r, _, err := g.client.Repositories.Get(context.Background(), g.owner, repo)
	if err != nil {
		return nil, fmt.Errorf("can't find repo %s of owner %s: %w", repo, g.owner, err)
	}

	return r, nil

}

func (g *Github) fetchCommitCount(repository string) (*github.RepositoryParticipation, error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	p, _, err := g.client.Repositories.ListParticipation(context.Background(), g.owner, repo)
	if err != nil {
		return nil, fmt.Errorf("can't find repo %s of owner %s: %w", repo, g.owner, err)
	}

	return p, nil
}

func (g *Github) fetchViews(repository string) (*github.TrafficViews, error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	t, _, err := g.client.Repositories.ListTrafficViews(context.Background(), g.owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("can't find repo %s of owner %s: %w", repo, g.owner, err)
	}

	return t, nil
}

func (g *Github) fetchBranches(repository string, limit int) ([]*github.Branch, error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	opt := github.BranchListOptions{ListOptions: github.ListOptions{PerPage: limit}}
	bs, _, err := g.client.Repositories.ListBranches(context.Background(), g.owner, repo, &opt)
	if err != nil {
		return nil, fmt.Errorf("can't find branches of owner %s for repo %s: %w", g.owner, repo, err)
	}

	return bs, nil
}

// Possibility to add options to filter quite a lot
func (g *Github) fetchIssues(repository string, limit int) ([]*github.Issue, error) {
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	opt := github.IssueListByRepoOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: limit},
	}
	is, _, err := g.client.Issues.ListByRepo(context.Background(), g.owner, repo, &opt)
	if err != nil {
		return nil, fmt.Errorf("can't find the issues of owner %s for repo %s: %w", g.owner, repo, err)
	}

	return is, nil
}

func (g *Github) fetchPullRequests(repository string, limit int) ([]*github.PullRequest, error) {
	ctx := context.Background()
	repo := g.repoName
	if repository != "" {
		repo = repository
	}

	if repo == "" {
		return nil, errors.New("you need to specify a repository in the github service or in the widget")
	}

	prs, _, err := g.client.PullRequests.List(ctx, g.owner, repo, &github.PullRequestListOptions{
		State:       "all",
		ListOptions: github.ListOptions{PerPage: limit},
	})
	if err != nil {
		return nil, fmt.Errorf("can't find the pull requests of owner %s for repo %s: %w", g.owner, repo, err)
	}

	return prs, nil
}

func (g *Github) fetchAllRepo(order string) ([]*github.Repository, error) {
	ctx := context.Background()

	r, _, err := g.client.Repositories.ListByUser(ctx, g.owner, &github.RepositoryListByUserOptions{Sort: order})
	if err != nil {
		return nil, fmt.Errorf("can't find all repo of owner %s: %w", g.owner, err)
	}

	return r, nil
}

func contains(slice []string, value string) bool {
	return slices.Contains(slice, value)
}

// fillMissingDays add the missing dates between two dates and add 0 values
func fillMissingDays(dates []time.Time, values []int) ([]time.Time, []int) {
	d := []time.Time{}
	val := []int{}
	for k, v := range dates {
		d = append(d, v)
		val = append(val, values[k])

		if len(dates) <= k+1 {
			return d, val
		}

		nextDate := dates[k+1]
		mg := missingDays(v, nextDate)
		for range mg {
			val = append(val, 0)
		}
		d = append(d, mg...)
	}

	return d, val
}
