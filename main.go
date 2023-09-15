package main

import (
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/jordan-wright/email"
	"github.com/mmcdole/gofeed"
	"github.com/russross/blackfriday/v2"
	_ "modernc.org/sqlite"
)

func main() {
	bootstrapConfig()

	fp := gofeed.NewParser()
	writer := getWriter()
	displayWeather(writer)

	for _, feed := range myFeeds {
		parsedFeed := parseFeed(fp, feed.url, feed.limit)

		if parsedFeed == nil {
			continue
		}

		items := generateFeedItems(writer, parsedFeed, feed)
		if items != "" {
			writeFeed(writer, parsedFeed, items)
		}
	}

	markdown_file_name := mdPrefix + currentDate + mdSuffix + ".md"
	sendEmail(markdown_file_name)

	// Close the database connection after processing all the feeds
	defer db.Close()
}

func sendEmail(markdownFileName string) {
	if from_email == "" || email_password == "" {
		return
	}

	// read content from markdown file
	markdownFile, err := os.Open(filepath.Join(markdownDirPath, markdownFileName))
	if err != nil {
		log.Fatal(err)
	}
	defer markdownFile.Close()

	// convert the file into string
	markdownFileContent, err := ioutil.ReadAll(markdownFile)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new email message
	e := email.NewEmail()
	e.From = from_email
	e.To = []string{from_email}
	e.Subject = "Daily Digest"
	e.HTML = blackfriday.Run(markdownFileContent)

	// Connect to the SMTP server
	auth := smtp.PlainAuth("", from_email, email_password, "smtp.gmail.com")
	err = e.Send("smtp.gmail.com:587", auth)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Email sent successfully!")
}
