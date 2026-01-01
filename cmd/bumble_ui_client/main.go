package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/vd09-projects/swipeassist/ui/bumble"
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

	client, err := bumble.NewClient(bumble.ClientConfig{
		LoginURL:   *loginURL,
		Headless:   *headless,
		ControlURL: *control,
	})
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

	if err := client.ScreenshotPage(ctx, filepath.Join(*outDir, "01.png")); err != nil {
		panic(err)
	}

	// 3) Client-controlled: next + screenshot
	for i := 0; i < *totalN; i++ {
		// let animation settle (client controls this)
		time.Sleep(2 * time.Second)

		name := fmt.Sprintf("%02d.png", i+1)
		if err := client.ScreenshotPage(ctx, filepath.Join(*outDir, name)); err != nil {
			panic(err)
		}

		if err := client.NextPicture(ctx); err != nil {
			fmt.Println("NextPicture stopped:", err)
			break
		}
	}

	// 4) Sleep before clicking action (as requested)
	time.Sleep(20 * time.Second)

	var act bumble.Action
	switch *actionStr {
	case "PASS":
		act = bumble.ActionPass
	case "LIKE":
		act = bumble.ActionLike
	case "SUPERSWIPE":
		act = bumble.ActionSuperSwipe
	default:
		panic("invalid -action. Use PASS | LIKE | SUPERSWIPE")
	}

	if err := client.ClickAction(ctx, act); err != nil {
		panic(err)
	}

	fmt.Println("Done.")
}
