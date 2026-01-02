package bumble

type Selectors struct {
	NextImage         []string
	NextImageDisabled []string
	Pass              []string
	SuperSwipe        []string
	Like              []string
	ReadyHints        []string
	AlbumNav          string
}

func DefaultSelectors() Selectors {
	return Selectors{
		NextImage: []string{
			"div.encounters-album__nav-item.encounters-album__nav-item--next[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div:nth-child(1) > article > div.encounters-album__nav > div.encounters-album__nav-item.encounters-album__nav-item--next",
		},
		NextImageDisabled: []string{
			"div.encounters-album__nav-item.is-disabled.encounters-album__nav-item--next[role='button']",
			"#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div:nth-child(1) > article > div.encounters-album__nav > div.encounters-album__nav-item.is-disabled.encounters-album__nav-item--next",
		},
		Pass: []string{
			"div[data-qa-role='encounters-action-dislike'][role='button']",
			"div.encounters-action.encounters-action--dislike[role='button']",
		},
		SuperSwipe: []string{
			"div[data-qa-role='encounters-action-superswipe'][role='button']",
			"div.encounters-action.encounters-action--superswipe[role='button']",
		},
		Like: []string{
			"div[data-qa-role='encounters-action-like'][role='button']",
			"div.encounters-action.encounters-action--like[role='button']",
		},
		ReadyHints: []string{
			"div.encounters-user__controls",
			"article",
		},
		AlbumNav: "#main > div > div.page__layout > main > div.page__content-inner > div > div > span > div:nth-child(1) > article > div.encounters-album__nav",
	}
}
