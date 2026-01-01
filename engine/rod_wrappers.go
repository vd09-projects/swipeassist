package engine

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Browser interface {
	MustPage(url string) Page
	MustClose()
}

type Page interface {
	MustWaitLoad() Page
	Timeout(d time.Duration) Page
	Element(selector string) (Element, error)
	Screenshot(fullPage bool, opt *proto.PageCaptureScreenshot) ([]byte, error)
	MustClose()
}

type Element interface {
	MustClick()
	ScrollIntoView() error
	EvalBool(js string) (bool, error)
	Screenshot(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error)
}

type RodBrowser struct{ Inner *rod.Browser }

func (b RodBrowser) MustPage(url string) Page { return RodPage{Inner: b.Inner.MustPage(url)} }
func (b RodBrowser) MustClose()               { b.Inner.MustClose() }

type RodPage struct{ Inner *rod.Page }

func (p RodPage) MustWaitLoad() Page { p.Inner.MustWaitLoad(); return p }
func (p RodPage) Timeout(d time.Duration) Page {
	return RodPage{Inner: p.Inner.Timeout(d)}
}
func (p RodPage) Element(selector string) (Element, error) {
	el, err := p.Inner.Element(selector)
	if err != nil || el == nil {
		return nil, err
	}
	return RodElement{Inner: el}, nil
}
func (p RodPage) Screenshot(fullPage bool, opt *proto.PageCaptureScreenshot) ([]byte, error) {
	return p.Inner.Screenshot(fullPage, opt)
}
func (p RodPage) MustClose() { _ = p.Inner.Close() }

type RodElement struct{ Inner *rod.Element }

func (e RodElement) MustClick()            { e.Inner.MustClick() }
func (e RodElement) ScrollIntoView() error { return e.Inner.ScrollIntoView() }

func (e RodElement) Screenshot(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error) {
	return e.Inner.Screenshot(format, quality)
}

func (e RodElement) EvalBool(js string) (bool, error) {
	obj, err := e.Inner.Eval(js)
	if err != nil {
		return false, err
	}
	if obj == nil {
		return false, nil
	}
	if obj.UnserializableValue != "" {
		return obj.UnserializableValue == "true", nil
	}
	return obj.Value.Bool(), nil
}
