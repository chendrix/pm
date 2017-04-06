package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/chendrix/pm/lib/gh"
	"github.com/chendrix/pm/lib/tablewriter"
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"github.com/vito/twentythousandtonnesofcrudeoil"
	"golang.org/x/oauth2"
)

type PassengerManifestCommand struct {
	GitHub struct {
		Token            string `long:"token"             required:"true" description:"GitHub access token"`
		OrganizationName string `long:"organization-name" required:"true" description:"GitHub organization name"`
	} `group:"GitHub Configuration" namespace:"github"`

	Debug bool `long:"debug" description:"Run in debug mode"`
}

func main() {
	cmd := &PassengerManifestCommand{}

	ctx := context.Background()

	parser := flags.NewParser(cmd, flags.Default)
	parser.NamespaceDelimiter = "-"

	twentythousandtonnesofcrudeoil.TheEnvironmentIsPerfectlySafe(parser, "PASSENGERMANIFEST_")

	args, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	w := os.Stdout
	logLevel := lager.INFO
	if cmd.Debug {
		logLevel = lager.DEBUG
	}

	logger := lager.NewLogger("passengermanifest")
	logger.RegisterSink(lager.NewWriterSink(w, logLevel))

	err = cmd.Execute(ctx, logger, os.Stdout, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func (cmd *PassengerManifestCommand) Execute(ctx context.Context, l lager.Logger, w io.Writer, argv []string) error {
	logger := l.Session("execute")

	ghToken := &oauth2.Token{AccessToken: cmd.GitHub.Token}

	ghAuth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(ghToken))

	githubClient := github.NewClient(ghAuth)

	ghClient := gh.NewClient(githubClient)

	logger.Debug("gathering issues")
	issues, err := ghClient.AllIssuesForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	logger.Debug("gathering issue comments")
	issueComments, err := ghClient.AllIssueCommentsForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	logger.Debug("gathering repository comments")
	repositoryComments, err := ghClient.AllRepositoryCommentsForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	t := tablewriter.NewCSVTableWriter(w)

	logger.Debug("calculating report")
	return Report(ctx, t, issues, issueComments, repositoryComments)
}

func Report(ctx context.Context, t tablewriter.TableWriter, issues []*github.Issue, issueComments []*github.IssueComment, repositoryComments []*github.RepositoryComment) error {
	u := NewUserList()

	for _, i := range issues {
		u.CatalogIssue(i)
	}

	for _, ic := range issueComments {
		u.CatalogIssueComment(ic)
	}

	for _, rc := range repositoryComments {
		u.CatalogRepositoryComment(rc)
	}

	t.SetHeader([]string{"Github User", "Opened Issues", "Issue Comments", "Repository Comments"})

	for name, user := range u {
		t.Append([]string{name, fmt.Sprintf("%d", len(user.OpenedIssues)), fmt.Sprintf("%d", len(user.IssueComments)), fmt.Sprintf("%d", len(user.RepositoryComments))})
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

func (u UserList) CatalogIssueComment(c *github.IssueComment) {
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

	user.AddIssueComment(c)
	u[*c.User.Login] = user
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
	IssueComments      []*github.IssueComment
	RepositoryComments []*github.RepositoryComment
}

func (u *User) AddOpenedIssue(i *github.Issue) {
	u.OpenedIssues = append(u.OpenedIssues, i)
}

func (u *User) AddIssueComment(c *github.IssueComment) {
	u.IssueComments = append(u.IssueComments, c)
}

func (u *User) AddRepositoryComment(c *github.RepositoryComment) {
	u.RepositoryComments = append(u.RepositoryComments, c)
}
