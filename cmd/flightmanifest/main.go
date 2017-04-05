package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/chendrix/pm/lib/gh"
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"github.com/olekukonko/tablewriter"
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

	issues, err := ghClient.AllIssuesForOrganization(ctx, cmd.GitHub.OrganizationName)
	if err != nil {
		return err
	}

	t := tablewriter.NewWriter(w)

	Report(ctx, t, issues)

	return nil
}

type TableWriter interface {
	SetHeader(keys []string)
	SetFooter(keys []string)
	Append(row []string)
	Render()
}

func Report(ctx context.Context, t TableWriter, issues []*github.Issue) {

	u := NewUserList()

	for _, i := range issues {
		u.AddUserFromIssue(i)
	}

	t.SetHeader([]string{"Github User", "Opened Issues"})

	for name, user := range u {
		t.Append([]string{name, fmt.Sprintf("%d", len(user.OpenedIssues))})
	}

	t.Render()
}

type UserList map[string]*User

func NewUserList() UserList {
	return make(map[string]*User)
}

func (u UserList) AddUserFromIssue(i *github.Issue) {
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

type User struct {
	GithubUser   *github.User
	OpenedIssues []*github.Issue
}

func (u *User) AddOpenedIssue(i *github.Issue) {
	u.OpenedIssues = append(u.OpenedIssues, i)
}
