package main

import (
	"embed"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/spf13/cobra"
)

//lint:ignore U1000 go:embed template.html
var htmlTemplate embed.FS

var rootCmd = &cobra.Command{
	Use:   "tinyfeed [FEED_URL ...]",
	Short: "Aggregate a collection of feed into static HTML page",
	Long:  "Aggregate a collection of feed into static HTML page. Only RSS, Atom and JSON feeds are supported.",
	Example: `  single feed      tinyfeed lovergne.dev/rss.xml > index.html
  multiple feeds   cat feeds.txt | tinyfeed > index.html`,
	Args: cobra.ArbitraryArgs,
	Run:  tinyfeed,
}

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
		fmt.Fprintln(os.Stderr, "you must input at list one feed url.")
		return
	}

	feeds := []*gofeed.Feed{}
	items := []*gofeed.Item{}
	fp := gofeed.NewParser()
	for _, url := range args {
		feed, err := fp.ParseURL(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not parse feed: %s\n", err)
			continue
		}
		feeds = append(feeds, feed)
		items = append(items, feed.Items...)
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].PublishedParsed.After(*items[j].PublishedParsed)
	})

	items = items[0:min(len(items), 49)]

	err = printHTML(feeds, items)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func Preview(item *gofeed.Item) string {
	if len(item.Description) > 0 {
		return truncstr(item.Description, 600)
	} else {
		return truncstr(item.Content, 600)
	}
}

func Domain(item *gofeed.Item) string {
	url, err := url.Parse(item.Link)
	if err != nil {
		log.Fatal(err)
	}
	hostname := strings.TrimPrefix(url.Hostname(), "www.")
	return hostname
}

func printHTML(feeds []*gofeed.Feed, items []*gofeed.Item) error {
	ts, err := template.New("template.html").
		Funcs(template.FuncMap{"Preview": Preview, "Domain": Domain}).
		ParseFiles("template.html")
	if err != nil {
		return fmt.Errorf("error loading html template: %s", err)
	}

	data := struct {
		Items []*gofeed.Item
		Feeds []*gofeed.Feed
	}{
		Items: items,
		Feeds: feeds,
	}

	err = ts.Execute(os.Stdout, data)
	if err != nil {
		return fmt.Errorf("error rendering html template: %s", err)
	}

	return nil
}
