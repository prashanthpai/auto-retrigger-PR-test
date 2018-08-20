package main

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/github"
)

const (
	Username        = "prashanthpai"
	PR              = 1150
	Password        = ""
	RetriggerPhrase = "retest this please"
)

func deleteComments(prID int) {
	t := &github.BasicAuthTransport{
		Username: Username,
		Password: Password,
	}

	client := github.NewClient(t.Client())

	comments, _, err := client.Issues.ListComments(context.Background(),
		"gluster", "glusterd2", prID, nil)
	if err != nil {
		log.Fatalf("failed to list comments on PR %d: Error: %s\n", prID, err.Error())
	}

	for _, comment := range comments {
		if strings.TrimSpace(comment.GetBody()) == RetriggerPhrase {
			log.Printf("Deleting comment with ID %d", comment.GetID())
			if _, err := client.Issues.DeleteComment(context.Background(),
				"gluster", "glusterd2", comment.GetID()); err != nil {
				log.Printf("failed to delete comment with ID %d: %s\n", comment.GetID(), err.Error())
			}
		}
	}
}

func main() {
	deleteComments(PR)
}
