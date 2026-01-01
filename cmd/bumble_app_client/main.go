package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/vd09-projects/swipeassist/apps"
	"github.com/vd09-projects/swipeassist/apps/bumble"
	"github.com/vd09-projects/swipeassist/apps/domain"
	// CHANGE this to your actual module import path
)

func main() {
	var (
		loginURL  = flag.String("login-url", "https://bumble.com/app", "Bumble URL (ensure you are logged in)")
		outDir    = flag.String("out", "./out/bumble", "Directory for screenshots")
		headless  = flag.Bool("headless", false, "Run headless")
		control   = flag.String("remote-url", "", "Rod ControlURL (optional). If empty, launches a new browser")
		totalN    = flag.Int("totalImages", 2, "How many times Screenshot (client-controlled)")
		actionStr = flag.String("action", "LIKE", "PASS | LIKE | SUPERSWIPE")
	)
	flag.Parse()

	ctx := context.Background()

	registry := map[string]apps.Adapter{
		"bumble": bumble.NewAdapterFromDefaults(),
	}
	client, err := apps.New(apps.Config{
		AppName:    "bumble",
		EntryURL:   *loginURL,
		Headless:   *headless,
		ControlURL: *control,
	}, registry)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// 1) Open
	if err := client.Open(ctx); err != nil {
		panic(err)
	}

	// // 2) Click picture (optional) + screenshot
	// if err := client.ClickPicture(ctx); err != nil {
	// 	fmt.Println("ClickPicture warning:", err)
	// }

	// if err := client.ScreenshotPage(ctx, filepath.Join(*outDir, "01.png")); err != nil {
	// 	panic(err)
	// }

	// 3) Client-controlled: next + screenshot
	for i := 0; i < *totalN; i++ {
		// let animation settle (client controls this)
		time.Sleep(2 * time.Second)

		name := fmt.Sprintf("%02d.png", i+1)
		if err := client.Screenshot(ctx, filepath.Join(*outDir, name)); err != nil {
			panic(err)
		}

		if err := client.NextMedia(ctx); err != nil {
			fmt.Println("NextPicture stopped:", err)
			break
		}
	}

	// 4) Sleep before clicking action (as requested)
	time.Sleep(20 * time.Second)

	var act domain.Action
	switch *actionStr {
	case "PASS":
		act.AType = domain.ActionPass
	case "LIKE":
		act.AType = domain.ActionLike
	case "SUPERSWIPE":
		act.AType = domain.ActionSuperSwipe
	default:
		panic("invalid -action. Use PASS | LIKE | SUPERSWIPE")
	}

	if err := client.Act(ctx, act); err != nil {
		panic(err)
	}

	fmt.Println("Done.")
}
