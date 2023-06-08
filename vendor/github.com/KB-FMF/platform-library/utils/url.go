package utils

import (
	"fmt"
	"github.com/KB-FMF/platform-library"
	"net/url"
	"strings"
)

func UrlParse(base string, elem ...string) (string, *platform.Error) {
	if strings.HasSuffix(base, "/") {
		base = strings.TrimSuffix(sanitizeURI(base[:len(base)-1]), "/")
	}

	// Parse
	rawURL, err := url.Parse(base)
	if err != nil {
		return "", platform.FromGoErr(err)
	}

	// Scheme
	if rawURL.Scheme != "http" && rawURL.Scheme != "https" {
		return "", platform.FromGoErr(fmt.Errorf("invalid scheme"))
	}

	// Host
	host := rawURL.Hostname()
	if host == "" {
		return "", platform.FromGoErr(fmt.Errorf("empty host"))
	}

	if len(elem) > 0 {
		elem = append([]string{base}, elem...)
		rawURL.Path = strings.Join(elem, "/")
	} else {
		return base, nil
	}

	return rawURL.Path, nil
}

func sanitizeURI(uri string) string {
	// double slash `\\`, `//` or even `\/` is absolute uri for browsers and by redirecting request to that uri
	// we are vulnerable to open redirect attack. so replace all slashes from the beginning with single slash
	if len(uri) > 1 && (uri[0] == '\\' || uri[0] == '/') && (uri[1] == '\\' || uri[1] == '/') {
		uri = "/" + strings.TrimLeft(uri, `/\`)
	}
	return uri
}
