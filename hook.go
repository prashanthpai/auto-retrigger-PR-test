package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bndr/gojenkins"
	"github.com/google/go-github/github"
)

const (
	Username        = "prashanthpai"
	RetriggerPhrase = "retest this please"
	JenkinsJob      = "gluster_glusterd2"
	PR              = 1150

	// fill these
	Password      = ""
	WebhookSecret = ""
)

func retriggerTests(prID int) {
	t := &github.BasicAuthTransport{
		Username: Username,
		Password: Password,
	}

	client := github.NewClient(t.Client())

	msg := RetriggerPhrase
	comment := github.IssueComment{
		Body: &msg,
	}
	if _, _, err := client.Issues.CreateComment(context.Background(),
		"gluster", "glusterd2", prID, &comment); err != nil {
		log.Printf("failed to create comment on PR %d\n", prID)
	}
}

var prRE = regexp.MustCompile("https://github.com/gluster/glusterd2/pull/[0-9]*")

func getPRFromBuild(build *gojenkins.Build) (int, error) {
	prURL := prRE.FindString(build.Raw.Description.(string))
	tmp := strings.Split(prURL, "/")
	return strconv.Atoi(tmp[len(tmp)-1])
}

func getBuildID(url string) (int, error) {
	tmp := strings.Split(url, "/")
	return strconv.Atoi(tmp[len(tmp)-2])
}

func processStatusEvent(event *github.StatusEvent) {

	targetURL := event.GetTargetURL()
	if targetURL == "" {
		return
	}

	eventState := event.GetState()
	log.Printf("%s : %s\n", targetURL, eventState)

	switch eventState {
	case "pending":
		return
	case "failure":
		// check if failed for relevant PR
	case "success":
		// pass, do nothing
	default:
		return
	}

	jenkins := gojenkins.CreateJenkins(nil, "https://ci.centos.org")
	if _, err := jenkins.Init(); err != nil {
		log.Println("jenkins.init() failed")
		return
	}

	buildID, err := getBuildID(targetURL)
	if err != nil {
		log.Println("failed to get build id")
		return
	}

	build, err := jenkins.GetBuild(JenkinsJob, int64(buildID))
	if err != nil {
		log.Println("failed to get build")
		return
	}

	pr, err := getPRFromBuild(build)
	if err != nil {
		log.Println("failed to get PR from build")
		return
	}

	if pr == PR {
		if eventState == "failure" {
			fmt.Printf("Phew! It failed: %s\n", targetURL)
			os.Exit(0)
		}
		log.Printf("Retriggering test for PR %d\n", pr)
		retriggerTests(pr)
	} else {
		log.Printf("NOT retriggering test for PR %d\n", pr)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(WebhookSecret))
	if err != nil {
		log.Print(err)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Print(err)
		return
	}

	switch event.(type) {
	case *github.StatusEvent:
		event, _ := event.(*github.StatusEvent)
		processStatusEvent(event)
	default:
		log.Print("Unsupported event received")
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe("localhost:4567", nil))
}
