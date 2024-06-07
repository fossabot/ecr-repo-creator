package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// createECRRepository creates an ECR repository if it does not exist.
func createECRRepository(repoName, region string) error {
	// Create a new session using the default credentials and region from the environment.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	svc := ecr.New(sess)

	// Check if the repository exists
	describeInput := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{aws.String(repoName)},
	}

	_, err = svc.DescribeRepositories(describeInput)
	if err == nil {
		log.Printf("Repository %s already exists.\n", repoName)
		return nil
	}

	// If the repository does not exist, create it
	createInput := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	}

	_, err = svc.CreateRepository(createInput)
	if err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	log.Printf("Repository %s created successfully.\n", repoName)
	return nil
}

// extractRepoName strips the ECR URL to get the repository name.
func extractRepoName(ecrURL string) string {
	for strings.HasSuffix(ecrURL, "/") {
		ecrURL = ecrURL[:len(ecrURL)-1]
	}

	split_url := strings.Split(ecrURL, "/")

	last_part := split_url[len(split_url)-1]
	if strings.Contains(last_part, ":") {
		split_url[len(split_url)-1] = strings.Split(last_part, ":")[0]
	}

	if strings.Contains(split_url[0], ".amazonaws.com") {
		return strings.Join(split_url[1:], "/")
	}

	return strings.Join(split_url, "/")
}

func main() {
	region := flag.String("region", "eu-west-1", "AWS region to use")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("Usage: %s [--region region] <repository-url-or-name>\n", os.Args[0])
	}

	repoInput := flag.Arg(0)

	repoName := extractRepoName(repoInput)

	err := createECRRepository(repoName, *region)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
