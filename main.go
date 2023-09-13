package main

import (
	_ "embed"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/spf13/cobra"
)

//go:embed built-in
var builtInTemplate string

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func tinyfeed(cmd *cobra.Command, args []string) {
	strdinArgs, err := stdinToArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	args = append(args, strdinArgs...)

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No argument found, you must input at least one feed url.")
		return
	}

	feeds := []*gofeed.Feed{}
	items := []*gofeed.Item{}
	fp := gofeed.NewParser()
	for _, url := range args {
		feed, err := fp.ParseURL(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse feed at %s: %s\n", url, err)
			continue
		}
		feeds = append(feeds, feed)
		items = append(items, feed.Items...)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PublishedParsed == nil {
			return false
		}
		if items[j].PublishedParsed == nil {
			return true
		}
		return items[i].PublishedParsed.After(*items[j].PublishedParsed)
	})

	items = items[0:min(len(items), limit)]

	err = printHTML(feeds, items)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func printHTML(feeds []*gofeed.Feed, items []*gofeed.Item) error {
	var err error
	var ts *template.Template

	if templatePath == "" {
		ts, err = template.New("built-in").
			Funcs(template.FuncMap{"preview": preview, "domain": domain}).
			Parse(builtInTemplate)
	} else {
		ts, err = template.New(templatePath).
			Funcs(template.FuncMap{"preview": preview, "domain": domain}).
			ParseFiles(templatePath)
	}
	if err != nil {
		return fmt.Errorf("error loading html template: %s", err)
	}

	imageCsp := "'none'"
	if imageAllowed {
		imageCsp = "https://*"
	}

	data := struct {
		Metadata map[string]string
		Items    []*gofeed.Item
		Feeds    []*gofeed.Feed
	}{
		Metadata: map[string]string{
			"name":       name,
			"imageCsp":   imageCsp,
			"stylesheet": stylesheet,
			"nonce":      randStr(20),
		},
		Items: items,
		Feeds: feeds,
	}

	err = ts.Execute(os.Stdout, data)
	if err != nil {
		return fmt.Errorf("error rendering html template: %s", err)
	}

	return nil
}
