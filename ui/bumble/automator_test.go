package bumble

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/go-rod/rod/lib/input"
// )

// func TestAutomatorOpenBumble(t *testing.T) {
// 	fp := &fakePage{mouse: &fakeMouse{}, keyboard: &fakeKeyboard{}}
// 	fb := &fakeBrowser{page: fp}
// 	a := &Automator{browser: fb}

// 	const loginURL = "https://example.com/login"
// 	if err := a.OpenBumble(context.Background(), loginURL); err != nil {
// 		t.Fatalf("OpenBumble returned error: %v", err)
// 	}
// 	if fb.lastURL != loginURL {
// 		t.Fatalf("expected login URL %q, got %q", loginURL, fb.lastURL)
// 	}
// 	if a.page != fp {
// 		t.Fatalf("expected automator page to be set")
// 	}
// }

// func TestAutomatorScrollAndClickImages(t *testing.T) {
// 	thumbs := []*fakeElement{{}, {}, {}}
// 	thumbElements := make([]element, len(thumbs))
// 	for i, thumb := range thumbs {
// 		thumbElements[i] = thumb
// 	}

// 	bigImages := []element{
// 		&fakeElement{attributes: map[string]string{"src": "img1"}},
// 		nil, // simulate missing enlarged image
// 		&fakeElement{attributes: map[string]string{"src": "img3"}},
// 	}

// 	mouse := &fakeMouse{}
// 	keyboard := &fakeKeyboard{}
// 	fp := &fakePage{
// 		elements:  thumbElements,
// 		bigImages: bigImages,
// 		mouse:     mouse,
// 		keyboard:  keyboard,
// 	}
// 	a := &Automator{page: fp}

// 	got, err := a.ScrollAndClickImages(context.Background())
// 	if err != nil {
// 		t.Fatalf("ScrollAndClickImages returned error: %v", err)
// 	}
// 	want := []string{"img1", "img3"}
// 	if len(got) != len(want) {
// 		t.Fatalf("unexpected image count: got %v want %v", got, want)
// 	}
// 	for i := range got {
// 		if got[i] != want[i] {
// 			t.Fatalf("unexpected images: got %v want %v", got, want)
// 		}
// 	}
// 	if fp.waitLoadCount != 10 {
// 		t.Fatalf("expected 10 wait loads, got %d", fp.waitLoadCount)
// 	}
// 	if len(mouse.scrollCalls) != 10 {
// 		t.Fatalf("expected 10 scroll calls, got %d", len(mouse.scrollCalls))
// 	}
// 	if len(fp.timeoutDurations) != len(thumbs) {
// 		t.Fatalf("expected timeout per thumb, got %d", len(fp.timeoutDurations))
// 	}
// 	for _, d := range fp.timeoutDurations {
// 		if d != 3*time.Second {
// 			t.Fatalf("expected 3s timeout, got %v", d)
// 		}
// 	}
// 	if len(fp.elementSelectors) != len(thumbs) {
// 		t.Fatalf("expected element lookup per thumb, got %d", len(fp.elementSelectors))
// 	}
// 	for _, sel := range fp.elementSelectors {
// 		if sel != "img.enlarged" {
// 			t.Fatalf("expected selector img.enlarged, got %s", sel)
// 		}
// 	}
// 	if len(keyboard.keys) != len(thumbs) {
// 		t.Fatalf("expected escape per thumb, got %d", len(keyboard.keys))
// 	}
// 	for _, k := range keyboard.keys {
// 		if k != input.Escape {
// 			t.Fatalf("expected escape key, got %v", k)
// 		}
// 	}
// 	for _, thumb := range thumbs {
// 		if thumb.clicks != 1 {
// 			t.Fatalf("expected thumb clicked once, got %d", thumb.clicks)
// 		}
// 	}
// }

// func TestAutomatorScrollAndClickImagesNoPage(t *testing.T) {
// 	a := &Automator{}
// 	if _, err := a.ScrollAndClickImages(context.Background()); err == nil {
// 		t.Fatalf("expected error when page is nil")
// 	}
// }

// func TestAutomatorClickAction(t *testing.T) {
// 	tests := []struct {
// 		action   string
// 		selector string
// 	}{
// 		{action: "PASS", selector: "button[data-testid='pass-button']"},
// 		{action: "LIKE", selector: "button[data-testid='like-button']"},
// 		{action: "SUPERSWIPE", selector: "button[data-testid='super-swipe-button']"},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.action, func(t *testing.T) {
// 			thumb := &fakeElement{}
// 			fp := &fakePage{mustElement: thumb}
// 			a := &Automator{page: fp}

// 			if err := a.ClickAction(tt.action); err != nil {
// 				t.Fatalf("ClickAction returned error: %v", err)
// 			}
// 			if len(fp.mustElementSelectors) == 0 || fp.mustElementSelectors[len(fp.mustElementSelectors)-1] != tt.selector {
// 				t.Fatalf("expected selector %s, got %v", tt.selector, fp.mustElementSelectors)
// 			}
// 			if thumb.clicks != 1 {
// 				t.Fatalf("expected thumb clicked once, got %d", thumb.clicks)
// 			}
// 		})
// 	}
// }

// func TestAutomatorClickActionUnsupported(t *testing.T) {
// 	fp := &fakePage{mustElement: &fakeElement{}}
// 	a := &Automator{page: fp}
// 	if err := a.ClickAction("BOGUS"); err == nil {
// 		t.Fatalf("expected error for unsupported action")
// 	}
// }

// // fakes implementing browser/page/mouse/keyboard/element

// type fakeBrowser struct {
// 	lastURL string
// 	page    page
// 	closed  bool
// }

// func (b *fakeBrowser) MustPage(url string) page {
// 	b.lastURL = url
// 	if b.page == nil {
// 		b.page = &fakePage{}
// 	}
// 	return b.page
// }

// func (b *fakeBrowser) MustClose() {
// 	b.closed = true
// }

// type fakePage struct {
// 	waitLoadCount         int
// 	mouse                 *fakeMouse
// 	keyboard              *fakeKeyboard
// 	elements              []element
// 	bigImages             []element
// 	timeoutDurations      []time.Duration
// 	elementSelectors      []string
// 	mustElementsSelectors []string
// 	mustElementSelectors  []string
// 	mustElement           element
// }

// func (p *fakePage) MustWaitLoad() page {
// 	p.waitLoadCount++
// 	return p
// }

// func (p *fakePage) Mouse() mouse {
// 	if p.mouse == nil {
// 		p.mouse = &fakeMouse{}
// 	}
// 	return p.mouse
// }

// func (p *fakePage) MustElements(selector string) []element {
// 	p.mustElementsSelectors = append(p.mustElementsSelectors, selector)
// 	return p.elements
// }

// func (p *fakePage) Timeout(d time.Duration) page {
// 	p.timeoutDurations = append(p.timeoutDurations, d)
// 	return p
// }

// func (p *fakePage) Element(selector string) (element, error) {
// 	p.elementSelectors = append(p.elementSelectors, selector)
// 	if len(p.bigImages) == 0 {
// 		return nil, nil
// 	}
// 	img := p.bigImages[0]
// 	p.bigImages = p.bigImages[1:]
// 	return img, nil
// }

// func (p *fakePage) MustElement(selector string) element {
// 	p.mustElementSelectors = append(p.mustElementSelectors, selector)
// 	if p.mustElement == nil {
// 		p.mustElement = &fakeElement{}
// 	}
// 	return p.mustElement
// }

// func (p *fakePage) Keyboard() keyboard {
// 	if p.keyboard == nil {
// 		p.keyboard = &fakeKeyboard{}
// 	}
// 	return p.keyboard
// }

// type fakeMouse struct {
// 	scrollCalls []struct {
// 		offsetX float64
// 		offsetY float64
// 		steps   int
// 	}
// }

// func (m *fakeMouse) Scroll(offsetX, offsetY float64, steps int) error {
// 	m.scrollCalls = append(m.scrollCalls, struct {
// 		offsetX float64
// 		offsetY float64
// 		steps   int
// 	}{offsetX: offsetX, offsetY: offsetY, steps: steps})
// 	return nil
// }

// type fakeKeyboard struct {
// 	keys []input.Key
// }

// func (k *fakeKeyboard) Press(key input.Key) error {
// 	k.keys = append(k.keys, key)
// 	return nil
// }

// type fakeElement struct {
// 	clicks     int
// 	attributes map[string]string
// }

// func (e *fakeElement) MustClick() {
// 	e.clicks++
// }

// func (e *fakeElement) Attribute(name string) (*string, error) {
// 	if e.attributes == nil {
// 		return nil, nil
// 	}
// 	if v, ok := e.attributes[name]; ok {
// 		val := v
// 		return &val, nil
// 	}
// 	return nil, nil
// }
