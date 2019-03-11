package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	var workers int
	var outputFile string
	app := cli.NewApp()
	app.Name = "sha12csv"
	app.Usage = "List the sha1 sum of files in a folder in a csv"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Print the files being processed",
		},
		cli.IntFlag{
			Name:        "workers, w",
			Usage:       "Set the number of workers",
			Value:       10,
			Destination: &workers,
		},
		cli.StringFlag{
			Name:        "name, n",
			Usage:       "The name of the output file, defaults to sha1sum.csv",
			Value:       "sha1sum",
			Destination: &outputFile,
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("V") {
			log.SetLevel(log.TraceLevel)
		}
		var files []string
		workQueue := make(chan string, workers)
		var wg sync.WaitGroup

		if err := filepath.Walk(c.Args().Get(0), visit(&files)); err != nil {
			return err
		}
		log.WithField("numFiles", len(files)).Trace("Done walking directory")
		output := make(chan string, len(files))
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(id int, queue chan string, output chan<- string, wg *sync.WaitGroup) {
				for path := range queue {
					file, err := os.Open(path)
					if err != nil {
						log.WithField("workerId", id).Error(err)
					}
					hash := sha1.New()
					if _, err := io.Copy(hash, file); err != nil {
						log.WithField("workerId", id).Error(err)
					}
					file.Close()
					hashInBytes := hash.Sum(nil)[:20]
					outSha1 := hex.EncodeToString(hashInBytes)
					log.WithFields(log.Fields{"workerId": id, "file": path, "sha1sum": outSha1}).Trace("Finished processing file")
					output <- fmt.Sprintf("%s, %s\n", path, outSha1)
				}
				wg.Done()
			}(i, workQueue, output, &wg)
		}

		for _, path := range files {
			workQueue <- path
		}
		close(workQueue)
		wg.Wait()
		close(output)
		var buffer bytes.Buffer
		for out := range output {
			buffer.WriteString(out)
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s.csv", outputFile), buffer.Bytes(), 0644); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, ".git/") {
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}
