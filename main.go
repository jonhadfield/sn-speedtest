package main

import (
	"flag"
	"fmt"
	"github.com/jonhadfield/gosn"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	iterations := flag.Int("iterations", 1, "times to run")
	purgeBeforeTests := flag.Bool("purge-before-tests", false, "purge items before running tests")
	purgePostIteration := flag.Bool("purge-post-iter", false, "purge after each iteration")
	flag.Parse()

	var authTimes []time.Duration
	var putTimes []time.Duration

	if *purgeBeforeTests {
		log.Println("Running initial purge")
		//start := time.Now()

		s, err := getSession()
		if err != nil {
			return err
		}
		if err := purgeItems(&s); err != nil {
			log.Fatal(err)
		}
		//elapsed := time.Since(start)
		//log.Printf("initial purge time: %s", elapsed)

	}

	for i := 0; i < *iterations; i++ {
		log.Println("ITERATION:", i+1)
		// authenticate
		start := time.Now()

		s, err := getSession()
		if err != nil {
			return err
		}
		elapsed := time.Since(start)
		log.Printf("Auth time: %s", elapsed)
		authTimes = append(authTimes, elapsed)

		items := generateItems(s.Mk, s.Ak)

		// put items
		start = time.Now()
		_, err = gosn.PutItems(gosn.PutItemsInput{
			Items:   items,
			Session: s,
		})
		elapsed = time.Since(start)
		log.Printf("Put time: %s", elapsed)
		putTimes = append(putTimes, elapsed)

		if *purgePostIteration {
			//start := time.Now()

			if err := purgeItems(&s); err != nil {
				log.Fatal(err)
			}
			//elapsed := time.Since(start)
			//log.Printf("post iteration purge time: %s", elapsed)
		}
	}

	// average the times
	var totalAuthTime time.Duration
	for _, at := range authTimes {
		totalAuthTime += at
	}
	var totalPutTime time.Duration
	for _, pt := range putTimes {
		totalPutTime += pt
	}
	fmt.Println("-----------")
	fmt.Printf("Average Auth time: %s\n", time.Duration(int64(totalAuthTime)/int64(*iterations)))
	fmt.Printf("Average Put time:  %s\n", time.Duration(int64(totalPutTime)/int64(*iterations)))

	return nil
}

func purgeItems(session *gosn.Session) (err error) {
	gnf := gosn.Filter{
		Type: "Note",
	}
	gtf := gosn.Filter{
		Type: "Tag",
	}
	f := gosn.ItemFilters{
		Filters:  []gosn.Filter{gnf, gtf},
		MatchAny: true,
	}
	gii := gosn.GetItemsInput{
		Session: *session,
	}

	var gio gosn.GetItemsOutput

	gio, err = gosn.GetItems(gii)
	if err != nil {
		return
	}

	var di gosn.DecryptedItems

	di, err = gio.Items.Decrypt(session.Mk, session.Ak)
	if err != nil {
		return
	}

	var items gosn.Items

	items, err = di.Parse()
	if err != nil {
		return
	}

	items.Filter(f)

	var toDel gosn.Items

	for x := range items {
		md := items[x]
		switch md.ContentType {
		case "Note":
			md.Content = gosn.NewNoteContent()
		case "Tag":
			md.Content = gosn.NewTagContent()
		}

		md.Deleted = true
		toDel = append(toDel, md)
	}

	if len(toDel) > 0 {
		eToDel, _ := toDel.Encrypt(session.Mk, session.Ak)
		putItemsInput := gosn.PutItemsInput{
			Session: *session,
			Items:   eToDel,
		}

		_, err = gosn.PutItems(putItemsInput)
		if err != nil {
			return fmt.Errorf("PutItems Failed: %v", err)
		}
	}

	return err
}

func generateItems(mk, ak string) gosn.EncryptedItems {
	// generate items to sync
	words := "these are some words"
	var items gosn.Items
	for j := 0; j < 100; j++ {
		noteContent := gosn.NewNoteContent()
		noteContent.Title = fmt.Sprintf("note title %d", j)
		noteContent.Text = fmt.Sprintf("%s", strings.Repeat(words, 300))
		note := gosn.NewNote()
		note.Content = noteContent
		items = append(items, *note)
	}

	// encrypt
	e, err := items.Encrypt(mk, ak)
	if err != nil {
		log.Fatal(err)
	}

	return e
}

func getSession() (session gosn.Session, err error) {
	defaultServer := "https://sync.standardnotes.org"
	email := os.Getenv("SN_EMAIL")
	if email == "" {
		err = fmt.Errorf("environment variable SN_EMAIL is required")
		return
	}

	password := os.Getenv("SN_PASSWORD")
	if password == "" {
		err = fmt.Errorf("environment variable SN_PASSWORD is required")
		return
	}

	server := os.Getenv("SN_SERVER")
	if server == "" {
		server = defaultServer
	}

	session, err = gosn.CliSignIn(email,
		password, server)

	return
}
