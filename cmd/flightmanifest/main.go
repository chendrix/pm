package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/chendrix/pm/lib/gh"
	"github.com/chendrix/pm/lib/tablewriter"
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"github.com/vito/twentythousandtonnesofcrudeoil"
	"golang.org/x/oauth2"
)

type FlightmanifestCommand struct {
	GitHub struct {
		Token            string `long:"token"             required:"true" description:"GitHub access token"`
		OrganizationName string `long:"organization-name" required:"true" description:"GitHub organization name"`
	} `group:"GitHub Configuration" namespace:"github"`
}

func main() {
	cmd := &FlightmanifestCommand{}

	ctx := context.Background()

	parser := flags.NewParser(cmd, flags.Default)
	parser.NamespaceDelimiter = "-"

	twentythousandtonnesofcrudeoil.TheEnvironmentIsPerfectlySafe(parser, "FLIGHTMANIFEST_")

	args, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	err = cmd.Execute(ctx, os.Stdout, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func (cmd *FlightmanifestCommand) Execute(ctx context.Context, w io.Writer, argv []string) error {
	ghToken := &oauth2.Token{AccessToken: cmd.GitHub.Token}

	ghAuth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(ghToken))

	githubClient := github.NewClient(ghAuth)

	ghClient := gh.NewClient(githubClient)

	log.Println("gathering issues")
	issues, err := ghClient.AllIssuesForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	log.Println("gathering repositoryComments")
	repositoryComments, err := ghClient.AllRepositoryCommentsForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	t := tablewriter.NewCSVTableWriter(w)

	log.Println("calculating report")
	return Report(ctx, t, issues, repositoryComments)
}

func Report(ctx context.Context, t tablewriter.TableWriter, issues []*github.Issue, repositoryComments []*github.RepositoryComment) error {

	u := NewUserList()

	for _, i := range issues {
		u.CatalogIssue(i)
	}

	for _, c := range repositoryComments {
		u.CatalogRepositoryComment(c)
	}

	t.SetHeader([]string{"Github User", "Opened Issues", "Repository Comments"})

	for name, user := range u {
		t.Append([]string{name, fmt.Sprintf("%d", len(user.OpenedIssues)), fmt.Sprintf("%d", len(user.RepositoryComments))})
	}

	return t.Render()
}

type UserList map[string]*User

func NewUserList() UserList {
	return make(map[string]*User)
}

func (u UserList) CatalogIssue(i *github.Issue) {
	var (
		user   *User
		exists bool
	)

	user, exists = u[*i.User.Login]
	if !exists {
		user = &User{
			GithubUser: i.User,
		}
	}

	user.AddOpenedIssue(i)
	u[*i.User.Login] = user
}

func (u UserList) CatalogRepositoryComment(c *github.RepositoryComment) {
	var (
		user   *User
		exists bool
	)

	user, exists = u[*c.User.Login]
	if !exists {
		user = &User{
			GithubUser: c.User,
		}
	}

	user.AddRepositoryComment(c)
	u[*c.User.Login] = user
}

type User struct {
	GithubUser         *github.User
	OpenedIssues       []*github.Issue
	RepositoryComments []*github.RepositoryComment
}

func (u *User) AddOpenedIssue(i *github.Issue) {
	u.OpenedIssues = append(u.OpenedIssues, i)
}

func (u *User) AddRepositoryComment(c *github.RepositoryComment) {
	u.RepositoryComments = append(u.RepositoryComments, c)
}
