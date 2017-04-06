package gh

import (
	"context"

	"github.com/google/go-github/github"
)

var publicReposFilter = github.RepositoryListByOrgOptions{Type: "public"}
var openIssuesFilter = github.IssueListByRepoOptions{State: "open"}

type Client struct {
	GithubClient *github.Client
}

func NewClient(githubClient *github.Client) *Client {
	return &Client{
		GithubClient: githubClient,
	}
}

func (client *Client) PublicRepositories(ctx context.Context, org string) ([]*github.Repository, error) {
	options := publicReposFilter

	var all []*github.Repository

	for {
		resources, resp, err := client.GithubClient.Repositories.ListByOrg(
			ctx,
			org,
			&options,
		)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			break
		}

		all = append(all, resources...)

		if resp.NextPage == 0 {
			break
		}

		options.ListOptions.Page = resp.NextPage
	}

	return all, nil
}

func (client *Client) AllIssuesForOrganization(ctx context.Context, org string) ([]*github.Issue, error) {
	repos, err := client.PublicRepositories(ctx, org)
	if err != nil {
		return nil, err
	}

	var all []*github.Issue
	for _, repo := range repos {
		issues, err := client.AllIssues(ctx, repo)
		if err != nil {
			return nil, err
		}

		all = append(all, issues...)
	}

	return all, nil
}

func (client *Client) AllRepositoryCommentsForOrganization(ctx context.Context, org string) ([]*github.RepositoryComment, error) {
	repos, err := client.PublicRepositories(ctx, org)
	if err != nil {
		return nil, err
	}

	var all []*github.RepositoryComment
	for _, repo := range repos {
		issues, err := client.AllCommentsForRepository(ctx, repo)
		if err != nil {
			return nil, err
		}

		all = append(all, issues...)
	}

	return all, nil
}

func (client *Client) AllIssues(ctx context.Context, repo *github.Repository) ([]*github.Issue, error) {
	options := openIssuesFilter

	var all []*github.Issue

	for {
		resources, resp, err := client.GithubClient.Issues.ListByRepo(
			ctx,
			*repo.Owner.Login,
			*repo.Name,
			&options,
		)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			break
		}

		all = append(all, resources...)

		if resp.NextPage == 0 {
			break
		}

		options.ListOptions.Page = resp.NextPage
	}

	return all, nil
}

func (client *Client) AllCommentsForRepository(
	ctx context.Context,
	repo *github.Repository,
) ([]*github.RepositoryComment, error) {
	options := &github.ListOptions{}

	var all []*github.RepositoryComment

	for {
		resources, resp, err := client.GithubClient.Repositories.ListComments(
			ctx,
			*repo.Owner.Login,
			*repo.Name,
			options,
		)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			break
		}

		all = append(all, resources...)

		if resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	return all, nil
}
