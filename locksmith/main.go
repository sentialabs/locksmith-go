package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/captainsafia/go-user-shell"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
	"gopkg.in/mattes/go-expand-tilde.v1"
)

const esc = "\033["

type Bookmarks struct {
	Links struct {
		Parent struct {
			Href string `json:"href"`
		} `json:"parent"`
		First struct {
			Href string `json:"href"`
		} `json:"first"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Last struct {
			Href string `json:"href"`
		} `json:"last"`
	} `json:"_links"`
	Bookmarks []struct {
		Links struct {
			Parent struct {
				Href string `json:"href"`
			} `json:"parent"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
		ID            int    `json:"id"`
		RoleName      string `json:"role_name"`
		Name          string `json:"name"`
		AccountNumber string `json:"account_number"`
		AvatarURL     string `json:"avatar_url"`
	} `json:"bookmarks"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

func warn(txt string) string {
	return fmt.Sprintf("%s%dm%s%s0m", esc, 91, txt, esc)
}

// Version number
var Version string

// Build (git hash)
var Build string

func version() {
	fmt.Printf("Locksmith %s (%s)", Version, Build)
	os.Exit(0)
}

func main() {
	inceptionPtr := flag.Bool("inception", false, "allow locksmith in locksmith")
	versionPtr := flag.Bool("version", false, "show version")
	flag.Parse()

	if *versionPtr {
		version()
	}

	if !*inceptionPtr && len(os.Getenv("AWS_SESSION_EXPIRES")) > 0 {
		fmt.Println(
			warn("Warning: ") +
				"You are running Locksmith from a shell that was spawned " +
				"from Locksmith itself. This is probably not what you want, exit " +
				"this shell and start Locksmith again. If you indeed intended to run " +
				"Locksmith using the currently assumed role, please use the " +
				"-inception argument.")
		os.Exit(41)
	}

	path, err := tilde.Expand("~/.aws/credentials")
	if err != nil {
		log.Fatal("tilde.Expand: ", err)
		return
	}

	cfg, err := ini.InsensitiveLoad(path)
	if err != nil {
		log.Fatal("ini.InsensitiveLoad: ", err)
		return
	}
	mfaSerial := cfg.Section("locksmith").Key("mfa_serial").String()
	url := cfg.Section("locksmith").Key("beagle_url").String()
	pass := cfg.Section("locksmith").Key("beagle_pass").String()

	fmt.Printf("Locksmith GO\n")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("http.NewRequest: ", err)
		return
	}

	req.SetBasicAuth("n/a", pass)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("client.Do: ", err)
		return
	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record Bookmarks

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
	}

	sort.Slice(record.Bookmarks, func(i, j int) bool {
		return record.Bookmarks[i].Name < record.Bookmarks[j].Name
	})

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▸ {{ .AccountNumber | red }}: {{ .Name | yellow }}",
		Inactive: "  {{ .AccountNumber | red | faint }}: {{ .Name | yellow | faint }}",
		Selected: "{{ .AccountNumber | red }}: {{ .Name | yellow }}",
		// 		Details: `
		// --------- Account ----------
		// {{ "Name:" | faint }}	{{ .Name }}
		// {{ "AccountNumber:" | faint }}	{{ .AccountNumber }}
		// {{ "RoleName:" | faint }}	{{ .RoleName }}`,
	}

	searcher := func(input string, index int) bool {
		bookmark := record.Bookmarks[index]
		name := strings.Replace(strings.ToLower(bookmark.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		name += bookmark.AccountNumber

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "AWS Account",
		Items:     record.Bookmarks,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	result, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	validate := func(input string) error {
		match, err := regexp.MatchString("^[0-9]{6}$", input)
		if err != nil {
			return err
		}

		if !match {
			return errors.New("Token must be 6 digits")
		}

		return nil
	}

	mfaPrompt := promptui.Prompt{
		Label:    "MFA Token",
		Validate: validate,
	}

	token, err := mfaPrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	svc := sts.New(session.Must(session.NewSessionWithOptions(session.Options{
		Profile: "locksmith",
	})))
	input := &sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(3600),
		RoleArn: aws.String(fmt.Sprintf(
			"arn:aws:iam::%s:role/%s",
			record.Bookmarks[result].AccountNumber,
			record.Bookmarks[result].RoleName)),
		RoleSessionName: aws.String("AssumeRoleSession"),
		SerialNumber:    aws.String(mfaSerial),
		TokenCode:       aws.String(token),
	}

	assumedRole, err := svc.AssumeRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case sts.ErrCodePackedPolicyTooLargeException:
				fmt.Println(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
			case sts.ErrCodeRegionDisabledException:
				fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	shell := os.Getenv("LOCKSMITH_SHELL")

	if len(shell) == 0 {
		shell = user_shell.GetUserShell()
	}

	cmd := exec.Command(shell, "-l")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", aws.StringValue(assumedRole.Credentials.AccessKeyId)),
		fmt.Sprintf("AWS_ASSUMED_ROLE_ARN=%s", aws.StringValue(input.RoleArn)),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", aws.StringValue(assumedRole.Credentials.SecretAccessKey)),
		fmt.Sprintf("AWS_SECURITY_TOKEN=%s", aws.StringValue(assumedRole.Credentials.SessionToken)),
		fmt.Sprintf("AWS_SESSION_ACCOUNT_ID=%s", record.Bookmarks[result].AccountNumber),
		fmt.Sprintf("AWS_SESSION_ACCOUNT_NAME=%s", record.Bookmarks[result].Name),
		fmt.Sprintf("AWS_SESSION_EXPIRES=%d", aws.TimeValue(assumedRole.Credentials.Expiration).Unix()),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", aws.StringValue(assumedRole.Credentials.SessionToken)),
		fmt.Sprintf("AWS_SESSION_USER_ARN=%s", aws.StringValue(assumedRole.AssumedRoleUser.Arn)),
		fmt.Sprintf("AWS_SESSION_USER_ID=%s", aws.StringValue(assumedRole.AssumedRoleUser.AssumedRoleId)),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdStartErr := cmd.Start()
	if cmdStartErr != nil {
		log.Fatal(cmdStartErr)
	}
	cmd.Wait()
}
