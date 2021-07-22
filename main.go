package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/tcnksm/go-input"
	"golang.org/x/oauth2"

	"github.com/carlisia/ghinfo/analytics"
	"github.com/carlisia/ghinfo/github"
)

const baseURL = "https://api.github.com"
const userAgent = "https://github.com/carlisia/ghinfo"

func main() {
	var input2, reportType string
	fmt.Print("Welcome! 🌞 Please choose a report kind...\n" +
		"Type 1 for the stargazers analytics.\n" +
		"Type 2 for the license types analytics.\n" +
		"$ ")
	fmt.Scanf("%s", &reportType)
	if reportType > "2" || reportType < "1" {
		fmt.Printf("Unfortunately %s is not an option. Please try again.\n", reportType)
		os.Exit(1)
	}
	fmt.Printf("Thank you, you have selected %s. We'll get your started.\n\n", reportType)

	const gitHubToken = "GH_TOKEN"
	userToken := os.Getenv(gitHubToken)
	if userToken == "" {
		fmt.Print("\nWhile I have your attention: it seems you don't have a personal " +
			"access token configured. Please be sure to set the enviroment variable `GH_Token` " +
			"with your personal token in order to authenticate and proceed.\n\n" +
			"👋")
		os.Exit(1)
	}

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	var err error
	var query, since, maxID string

	query = "What is the Min ID?"
	since, err = ui.Ask(query, &input.Options{
		Default:  "65624570",
		Required: true,
	})
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	sinceInt, err := strconv.Atoi(since)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	query = "What is the Max ID?"
	maxID, err = ui.Ask(query, &input.Options{
		Default:  "65624720",
		Required: true,
	})
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	maxIDInt, err := strconv.Atoi(maxID)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	var column string
	if reportType == analytics.StarGazersReportType {
		fmt.Print("Please choose a column to order by:\n" +
			"1- bucket\n" +
			"2- repository\n" +
			"3- total stars \n" +
			"$ ")
		fmt.Scanf("%s", &column)
		if column > "3" || column < "1" {
			fmt.Printf("Unfortunately %s is not an option. Please try again.\n", column)
			os.Exit(1)
		}
	} else {
		fmt.Print("Please choose a column to order by::\n" +
			"1- license type \n" +
			"2- repository \n" +
			"$ ")
		fmt.Scanf("%s", &column)
		if column > "2" || column < "1" {
			fmt.Printf("Unfortunately %s is not an option. Please try again.\n", column)
			os.Exit(1)
		}
	}

	var asc string
	fmt.Print("Please choose an order to sort by:\n" +
		"0- asc \n" +
		"1- desc \n" +
		"$ ")
	fmt.Scanf("%s", &asc)
	if asc > "1" || asc < "0" {
		fmt.Printf("Unfortunately %s is not an option. Please try again.\n", asc)
		os.Exit(1)
	}
	ascBool, err := strconv.ParseBool(asc)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	opts := analytics.ParamOptions(
		analytics.ParamOptions{
			Column: column,
			Asc:    !ascBool,
			Since:  sinceInt,
			MaxID:  maxIDInt,
		},
	)

	var report analytics.StatsReport
	report, err = analytics.NewReport(reportType, opts)
	if err != nil {
		log.Fatalln("Invalid options were selected:", err)
	}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: userToken,
		},
	)

	ctx := context.Background()
	client := oauth2.NewClient(ctx, tokenSource)
	gh, err := github.New(client, baseURL, userAgent)
	if err != nil {
		log.Fatalln("Error trying to initalize the GitHub client:", err)
	}

	fmt.Printf("\nWe are about to retrive data for your %s ...\n\n", report.Name())

	if err := report.Run(ctx, gh); err != nil {
		log.Fatalln("Error trying to retrieve the repository list:", err)
	}

	fmt.Printf("We have retrieved %d repositories for your report. Would you like to have a print out? Please type `n` to exit.\n"+
		"$ ", report.Count())
	fmt.Scanf("%s", &input2)
	if input2 == "n" {
		os.Exit(0)
	}
	fmt.Print("Proceeding........\n\n")

	report.PrintStats()
}
