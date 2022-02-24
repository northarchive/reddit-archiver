package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/viper"
	"github.com/wader/goutubedl"
	"io"
	"log"
	"os"
	"time"
)

func (rtData *RuntimeData) workQueue() {
	for {
		fmt.Println("[QMNGR] Working queue...")

		rtData.WorkWg.Add(5)
		go func() {
			rtData.doQueueStep()
			rtData.WorkWg.Done()
		}()
		go func() {
			rtData.doQueueStep()
			rtData.WorkWg.Done()
		}()
		go func() {
			rtData.doQueueStep()
			rtData.WorkWg.Done()
		}()
		go func() {
			rtData.doQueueStep()
			rtData.WorkWg.Done()
		}()
		go func() {
			rtData.doQueueStep()
			rtData.WorkWg.Done()
		}()

		rtData.WorkWg.Wait()
		fmt.Println("[QMNGR] Queue work done, sleeping...")

		time.Sleep(time.Second * 2)

		if rtData.WantsExit {
			return
		}
	}
}

func (rtData *RuntimeData) doQueueStep() {
	if len(rtData.Queue) == 0 {
		return
	}

	rtData.QueueMutex.Lock()
	selected := rtData.Queue[0]
	rtData.Queue = rtData.Queue[1:]
	rtData.QueueMutex.Unlock()

	log.Printf(`[QWRKR] Now working on %s`, selected.ID)

	dateTime := selected.Created.Format("2006-01-02 15_04_05")

	folderName := fmt.Sprintf("%s"+string(os.PathSeparator)+"%s - %s"+string(os.PathSeparator)+"%s - %s - %s", viper.GetString("output_dir"), sanitize.BaseName(selected.SubredditID), sanitize.BaseName(selected.SubredditName), sanitize.BaseName(dateTime), sanitize.BaseName(selected.ID), sanitize.BaseName(TruncateString(selected.Title, 10)))

	os.MkdirAll(folderName, os.ModePerm)

	encoded, _ := json.Marshal(selected)

	os.WriteFile(fmt.Sprintf("%s"+string(os.PathSeparator)+"post.json", folderName), encoded, os.ModePerm)

	if selected.URL != "" {
		link := selected.URL
		dlLink(folderName, link)
	}
	if selected.URL == "" {
		link := fmt.Sprintf("https://reddit.com%s", selected.Permalink)
		dlLink(folderName, link)
	}
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	truncated := ""
	count := 0
	for _, char := range str {
		truncated += string(char)
		count++
		if count >= length {
			break
		}
	}
	return truncated
}

func dlLink(folderName string, link string) {
	log.Printf("[QWRKR] [%s] Downloading...", link)
	result, err := goutubedl.New(context.Background(), link, goutubedl.Options{})
	if err != nil {
		if err.Error() == "No media found" {
			return
		}
		log.Printf("[QWRKR] [%s] %v", link, err)
		return
	}
	downloadResult, err := result.Download(context.Background(), "best")
	if err != nil {
		log.Printf("[QWRKR] [%s] %v", link, err)
		return
	}
	defer downloadResult.Close()

	ytDlFilePath := fmt.Sprintf("%s"+string(os.PathSeparator)+"media.%s", folderName, result.Info.Ext)

	f, err := os.Create(ytDlFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	io.Copy(f, downloadResult)

	log.Printf("[QWRKR] [%s] Video saved to %s", link, ytDlFilePath)
}
