package extraction

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var helloPrint = regexp.MustCompile(`^print\(["'](?i:hello)[, ]+(?i:world)[! .]*["']\)$`)

// PythonHello extracts an entire standalone statement, never a line detached
// from a larger program (which could depend on imports or surrounding state).
func PythonHello(raw []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	code := ""
	doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
		candidate := strings.TrimSpace(s.Text())
		if helloPrint.MatchString(candidate) {
			code = candidate
		}
	})
	return code, nil
}

func PythonPrintSummary(raw []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	return clean(doc.Find("#print").Parent().ChildrenFiltered("dd").ChildrenFiltered("p").First().Text()), nil
}
