package main

import (
	"fmt"
	"github.com/tomasen/rollover"
	"github.com/tomasen/selfupdater"
	"time"
)

func main() {
	_, err := rollover.Wait()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		updater := selfupdater.NewSelfUpdate(selfupdater.NewS3UpdateProvider(
			"selfupdater-tomasen",
			"us-east-1",
			"/selfupdater",
			"/selfupdater.md5",
		))
		err := updater.Update()
		if err != nil {
			fmt.Println("updater returns:", err)
		}
		fmt.Println("update finished")
	}()

	ticker := time.NewTicker(5 * time.Second)
	for {
		fmt.Println("this is the base version")
		<-ticker.C
	}
}
