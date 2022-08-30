package main

import (
	"archive/zip"
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/schoentoon/servarr-backup/pkg"
	"github.com/sirupsen/logrus"
)

func main() {
	client := pkg.Client{}
	flag.StringVar(&client.BaseURL, "baseurl", "", "Base url of the servarr")
	flag.StringVar(&client.APIKey, "apikey", "", "Api key for the servarr")
	output := flag.String("output", "-", "Where to output the zip file to (- is stdout)")
	extract := flag.Bool("extract", false, "Should we extract the zip file?")
	flag.Parse()

	if client.BaseURL == "" {
		logrus.Fatal("No base url specified")
	}
	if client.APIKey == "" {
		logrus.Fatal("No api key specified")
	}
	if *output == "-" && *extract {
		logrus.Fatal("Extract and output to stdout are mutually exclusive")
	}

	ctx := context.Background()

	backup, err := client.StartBackup(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	err = backup.Wait(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	zipFile, err := client.DownloadLatestBackup(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer zipFile.Close()

	if !*extract {
		outputFile := os.Stdout
		if *output != "-" {
			outputFile, err = os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		_, err = io.Copy(outputFile, zipFile)
		if err != nil {
			logrus.Fatal(err)
		}

		// as we don't have to extract, this is our final step
		return
	}

	// so as we have arrived here we have to extract the file

	tmpFile, err := os.CreateTemp("/tmp", "servarr.zip.*")
	if err != nil {
		logrus.Fatal(err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	n, err := io.Copy(tmpFile, zipFile)
	if err != nil {
		logrus.Fatal(err)
	}

	archive, err := zip.NewReader(tmpFile, n)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, f := range archive.File {
		filePath := filepath.Join(*output, f.Name)

		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(filePath, os.ModePerm); err != nil {
				logrus.Fatal(err)
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			logrus.Fatal(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			logrus.Fatal(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			logrus.Fatal(err)
		}

		if _, err = io.Copy(dstFile, fileInArchive); err != nil {
			logrus.Fatal(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
}