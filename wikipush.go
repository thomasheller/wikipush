package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sadbox/mediawiki"
)

func main() {
	const doneDir = "done"
	const dupeDir = "dupes"
	const skipDir = "skipped"

	run := flag.Bool("run", false, "actually perform the upload")
	extension := flag.String("ext", ".txt", "file extension (including dot)")
	url := flag.String("url", "", "API url (typically http://.../w/api.php)")
	pause := flag.Duration("pause", 500*time.Millisecond, "wait time between uploads")
	summary := flag.String("summary", "Bulk upload by wikipush", "message for the revision log")

	flag.Parse()

	files, err := filepath.Glob("*" + *extension)
	if err != nil {
		log.Fatalf("Error reading filenames: %s", err)
	}

	if !*run {
		fmt.Printf("%d files would be uploaded.\n", len(files))
		if len(files) > 0 {
			fmt.Println("Start wikipush with `-run' flag to perform the upload.")
		}
		return
	}

	if len(files) == 0 {
		fmt.Println("No files found.")
		return
	}

	if *url == "" {
		log.Fatal("Error: Missing URL: Please specify your MediaWiki API URL using the `-url' flag.")
	}

	checkDir(doneDir, "done")
	checkDir(dupeDir, "duplicates")
	checkDir(skipDir, "skipped")

	client, err := mediawiki.New(*url, "wikipush")

	if err != nil {
		log.Fatalf("Error connecting to MediaWiki: %s\n", err)
	}

	in := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _, err := in.ReadLine()

	if err != nil {
		log.Fatalf("Error reading username: %s\n", err)
	}

	fmt.Print("Enter password (will be echoed): ")
	password, _, err := in.ReadLine()

	if err != nil {
		log.Fatalf("Error reading password: %s\n", err)
	}

	err = client.Login(string(username), string(password))

	if err != nil {
		log.Fatalf("Error logging in (wrong username/password?): %s\n", err)
	}

	defer client.Logout()

	fmt.Printf("%d files will be uploaded.\n", len(files))

	duration := *pause * time.Duration(len(files))

	fmt.Printf("This will take at least %s with the current throttle setting.\n", duration)

	var doneCount, dupeCount, skipCount, errCount, moveErrCount int

	for _, file := range files {
		local, err := ioutil.ReadFile(file)

		if err != nil {
			log.Printf("Error reading file %s, skipping...\n", file)
			errCount++
			continue
		}

		title := strings.TrimSuffix(file, *extension)

		time.Sleep(*pause)

		page, err := client.Read(title)

		if err != nil {
			log.Printf("Error fetching page %s, skipping...\n", title)
			errCount++
			continue
		}

		if len(page.Revisions) == 0 {
			edit := map[string]string{
				"title":   title,
				"summary": *summary,
				"text":    string(local),
			}

			err = client.Edit(edit)

			if err != nil {
				log.Printf("Error uploading page %s, skipping...\n", title)
				errCount++
				continue
			}

			log.Printf("Successfully uploaded \"%s\", page didn't exist before.\n", title)

			doneCount++

			err := os.Rename(file, path.Join(doneDir, file))

			if err != nil {
				log.Printf("Error moving file \"%s\" into \"%s\" directory, skipping...\n", file, doneDir)
				moveErrCount++
			}
		} else {
			remote := page.Revisions[0].Body

			log.Printf("Number of revisions: %d\n", len(page.Revisions))

			if strings.TrimSpace(string(local)) == strings.TrimSpace(remote) {
				log.Printf("Skipped \"%s\", because it's already in the wiki.\n", title)

				skipCount++

				err := os.Rename(file, path.Join(skipDir, file))

				if err != nil {
					log.Printf("Error moving file \"%s\" into \"%s\" directory, skipping...\n", file, skipDir)
					moveErrCount++
				}

			} else {
				log.Printf("Duplicate page \"%s\" with different content.\n", title)

				dupeCount++

				err := os.Rename(file, path.Join(dupeDir, file))

				if err != nil {
					log.Printf("Error moving file \"%s\" into \"%s\" directory, skipping...\n", file, dupeDir)
					moveErrCount++
				}

			}
		}
	}

	fmt.Println("Done.")
	fmt.Printf("Out of %d files,\n", len(files))
	fmt.Printf("%d were uploaded successfully (see %s directory),\n", doneCount, doneDir)
	fmt.Printf("%d were skipped because a page with different content already existed (see %s directory),\n", dupeCount, dupeDir)
	fmt.Printf("%d were skipped because they were alreay in the wiki (see %s directory).\n", skipCount, skipDir)
	fmt.Printf("%d couldn't be processed because of errors.\n", errCount)
	if moveErrCount > 0 {
		fmt.Printf("Additionally, %d files couldn't be moved into the correct directory after they were processed.\n", moveErrCount)
	}
}

func checkDir(dir string, description string) {
	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		log.Fatalf("Error creating %s directory \"%s\": %s\n", description, dir, err)
	}

	isEmpty, err := isDirEmpty(dir)

	if err != nil {
		log.Fatalf("Error checking if %s directory \"%s\" is empty: %s\n", description, dir, err)
	}

	if !isEmpty {
		log.Fatalf("Error: Please make sure the %s directory \"%s\" is empty.\n", description, dir)
	}
}

func isDirEmpty(name string) (bool, error) {
	file, err := os.Open(name)

	if err != nil {
		return false, err
	}

	defer file.Close()

	_, err = file.Readdirnames(1)

	if err == io.EOF {
		return true, nil
	}

	return false, err
}
