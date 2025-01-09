package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"

	"demoGolangGitCreation/env"
)

var (
	prAzureToken = env.Var("AZURE_DEVOPS_EXT_PAT").FallsbackTo(env.Var("AZURE_DEVOPS_TOKEN"))
	azureURL     = "https://dev.azure.com/tanzu-scc/catalog_test"
)

// AzureHost Represents an Azure Devops client config and information for the Azure Devops target
type AzureHost struct {
	client   *scm.Client
	host     string
	org      string
	token    string
	username string
	ca_cert  string
}

func main() {
	err, host := createClient()
	timedName := getTimedName()

	repo, err := host.CreateRepository(context.TODO(), timedName)
	if err != nil {
		panic(fmt.Errorf("could not create azure repository '%s': %w", timedName, err))
	}

	fmt.Printf("Success! Visit %s/%s\n", host.host, repo.FullName)
}

func createClient() (error, *AzureHost) {
	var fullOrg *url.URL
	var err error
	fullOrg, err = url.Parse(azureURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse url '%s': %w", azureURL, err))
	}

	host := &AzureHost{
		token:    prAzureToken.MustResolve(),
		host:     fullOrg.Scheme + "://" + fullOrg.Host,
		org:      strings.TrimPrefix(fullOrg.Path, "/"),
		username: "_token",
	}

	client, err := factory.NewClient("azure", host.host, host.token)
	host.client = client
	if err != nil {
		panic(fmt.Errorf("could not create azure client: '%w'", err))
	}
	return err, host
}

func getTimedName() string {
	dateTime := time.Now().Format(time.DateTime)

	// regular expression to match punctuation and spaces
	re := regexp.MustCompile(`[[:punct:]\s]`)
	// Replace all matches with underscores
	return re.ReplaceAllString(dateTime, "_")
}

func (h AzureHost) CreateRepository(ctx context.Context, name string) (*scm.Repository, error) {
	var err error
	var repo *scm.Repository
	maxRetries := 2

	ri := &scm.RepositoryInput{
		Namespace:   h.org,
		Name:        name,
		Description: "catalog gitops test repository",
		Private:     true,
	}

	for retries := 0; retries < maxRetries; retries++ {
		repo, _, err = h.client.Repositories.Create(ctx, ri)

		if err == nil {
			break
		}
		fmt.Printf(fmt.Errorf("create repo failed, will retry. retryError: %w, repository_name: %s\n", err, name).Error())
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	return repo, nil
}
