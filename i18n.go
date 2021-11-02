package mux

import (
	"net/http"

	"golang.org/x/text/language"
)

// localeMatcher is a wrapper around x/text/language#Matcher.
type localeMatcher struct {
	tags    []language.Tag
	matcher language.Matcher
}

// newLocaleMatcher returns a new locale matcher for the supported tags.
func newLocaleMatcher(tags []language.Tag) *localeMatcher {
	return &localeMatcher{
		tags:    tags,
		matcher: language.NewMatcher(tags),
	}
}

// match parses the Accept-Language header and returns
// the supported BCP 47 language tag for the request.
func (l *localeMatcher) match(req *http.Request) language.Tag {
	accept := req.Header.Get("Accept-Language")
	// Intentionally ignored error as the default language will be matched.
	tags, _, _ := language.ParseAcceptLanguage(accept)
	// https://github.com/golang/go/issues/24211
	_, i, _ := l.matcher.Match(tags...)
	return l.tags[i]
}
