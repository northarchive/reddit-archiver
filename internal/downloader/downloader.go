package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type RuntimeData struct {
	SeenIDs map[string]bool

	LastEntry map[string]string

	Queue []*reddit.Post

	QueueMutex sync.Mutex     `json:"-"`
	WorkWg     sync.WaitGroup `json:"-"`

	WantsExit bool `json:"-"`
}

func Run() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	saveFileName := fmt.Sprintf("%s/savefile.json", viper.GetString("output_dir"))

	client, _ := reddit.NewReadonlyClient()

	err := os.MkdirAll(viper.GetString("output_dir"), os.ModePerm)
	if err != nil {
		panic(err)
	}

	rtData := &RuntimeData{}

	f, err := os.Open(saveFileName)
	if err == nil {
		log.Println("Restoring save file...")
		json.NewDecoder(f).Decode(rtData)
	}

	if rtData.SeenIDs == nil {
		rtData.SeenIDs = make(map[string]bool)
	}
	if rtData.LastEntry == nil {
		rtData.LastEntry = make(map[string]string)
	}

	go rtData.workQueue()

	file, err := os.Open(viper.GetString("subreddit_list_file"))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	_, err = io.Copy(&buf, file)
	if err != nil {
		panic(err)
	}

	list := buf.String()
	entries := strings.Split(list, "\n")

	ticker := time.Tick(time.Second * 10)

	go func() {
		for {
			for _, subreddit := range entries {
				dlSubreddit(subreddit, client, rtData)
			}

			<-ticker

			if rtData.WantsExit {
				break
			}
		}
	}()

	<-c

	log.Println("Quit signal received, waiting for workers to finish...")

	rtData.WorkWg.Wait()
	log.Println("Shutting down...")

	saveFile, err := json.Marshal(rtData)
	if err != nil {
		log.Println(err)
		return
	}

	err = os.WriteFile(saveFileName, saveFile, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Saved state file.")
}

func dlSubreddit(subreddit string, client *reddit.Client, rtData *RuntimeData) error {
	if _, ok := rtData.LastEntry[subreddit]; !ok {
		rtData.LastEntry[subreddit] = ""
	}

	opts := reddit.ListOptions{
		Limit: 100,
		After: rtData.LastEntry[subreddit],
	}

	posts, _, err := client.Subreddit.NewPosts(context.TODO(), subreddit, &opts)
	if err != nil {
		return err
	}

	for i, post := range posts {
		if _, ok := rtData.SeenIDs[post.ID]; ok {
			continue
		}

		if i == 0 {
			rtData.LastEntry[subreddit] = post.ID
		}

		rtData.SeenIDs[post.ID] = true

		rtData.QueueMutex.Lock()

		rtData.Queue = append(rtData.Queue, post)

		rtData.QueueMutex.Unlock()

		log.Printf(`[WTCHR] New Post: %s`, post.Permalink)
	}

	return nil
}
