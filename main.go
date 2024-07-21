package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
)

// createECRRepository creates an ECR repository if it does not exist.
func createECRRepository(repoName, region string) error {
	// Create a new session using the default credentials and region from the environment.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)

	// Check if the repository exists
	describeInput := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{repoName},
	}

	_, err = ecrClient.DescribeRepositories(context.TODO(), describeInput)
	if err == nil {
		log.Printf("Repository %s already exists.\n", repoName)
		return nil
	} else {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			if ae.ErrorCode() != "RepositoryNotFoundException" {
				log.Printf("code: %s, message: %s, fault: %s", ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String())
				return err
			}
		} else {
			return err
		}
	}

	// If the repository does not exist, create it
	createInput := &ecr.CreateRepositoryInput{
		RepositoryName: aws.String(repoName),
	}

	_, err = ecrClient.CreateRepository(context.TODO(), createInput)
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
