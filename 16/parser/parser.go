package parser

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"wget/downloader"

	"golang.org/x/net/html"
)

type ParserHTML struct {
	maxDepth   int
	d          *downloader.Downloader
	host       string
	u          string
	urlToLocal map[string]string
	visitedURL map[string]struct{}
}

func normalizeURL(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Path == "" {
		u.Path = "/"
	} else if !strings.HasSuffix(u.Path, "/") && !strings.Contains(u.Path, ".") {
		u.Path += "/"
	}
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""

	return u.String(), nil
}

func NewParserHTML(maxDepth int, u string) ParserHTML {
	res, _ := url.Parse(u)
	return ParserHTML{
		visitedURL: make(map[string]struct{}),
		d:          &downloader.Downloader{},
		maxDepth:   maxDepth,
		u:          u,
		host:       res.Host,
		urlToLocal: make(map[string]string),
	}
}

func CheckHostUrl(u string) bool {
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		return false
	}
	return true
}

func (p *ParserHTML) checkValidUrl(u string) (bool, error) {
	if _, ok := p.visitedURL[u]; ok {
		return false, nil
	} else {
		p.visitedURL[u] = struct{}{}
	}

	uPath, err := url.Parse(u)
	if err != nil {
		return false, err
	}
	if uPath.Host != p.host && uPath.Host != "" {
		return false, nil
	}
	return true, nil
}

func (p *ParserHTML) Parse() error {
	err := p.parse(p.u, 0)
	if err != nil {
		return err
	}
	return nil
}

func (p *ParserHTML) parse(u string, currentDepth int) error {
	if currentDepth > p.maxDepth {
		return nil
	}
	normURL, err := normalizeURL(u)
	if err != nil {
		return err
	}
	ok, err := p.checkValidUrl(normURL)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	log.Println("Обработка ссылки: ", normURL)
	if !CheckHostUrl(normURL) {
		if !strings.Contains(normURL, p.host) {
			normURL = "https://" + p.host + "/" + normURL
		} else {
			normURL = "https://" + normURL
		}
	}

	res, err := http.Get(normURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	basePath := p.d.GetBasePath(normURL)

	filename, err := p.d.DownloadFile(basePath, p.d.GetPath(normURL), res.Body)
	if err != nil {

		return err
	}

	p.urlToLocal[normURL] = filename

	ext := strings.ToLower(filepath.Ext(filename))

	if ext == ".html" || ext == ".htm" {
		in, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer func() {
			if err := in.Close(); err != nil {
				log.Println(err)
			}
		}()

		links, err := p.FindAllLink(in)
		if err != nil {
			return err
		}
		err = p.UpdateLinks(filename)
		if err != nil {
			return err
		}
		if currentDepth+1 <= p.maxDepth {
			for _, l := range links {
				err := p.parse(l, currentDepth+1)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func (p *ParserHTML) FindAllLink(in io.ReadCloser) ([]string, error) {

	links := make([]string, 0)

	node, err := html.Parse(in)
	if err != nil {
		return nil, err
	}

	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					raw := strings.TrimSpace(attr.Val)
					if _, visited := p.visitedURL[raw]; !visited && raw != "" &&
						!strings.HasPrefix(raw, "javascript:") && !strings.HasPrefix(raw, "#") {
						u, err := url.Parse(raw)
						if err != nil {
							log.Println(err)
						}
						if u.Host == p.host || u.Host == "" {
							links = append(links, raw)
						}

					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(node)
	return links, nil
}

func (p *ParserHTML) UpdateLinks(htmlFilename string) error {
	f, err := os.Open(htmlFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	doc, err := html.Parse(f)
	if err != nil {
		return err
	}

	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				key := n.Attr[i].Key
				val := n.Attr[i].Val
				if key == "href" || key == "src" {
					absURL, err := url.Parse(val)
					if err != nil || absURL.Scheme == "" {
						baseURL, _ := url.Parse(p.u)
						absURL = baseURL.ResolveReference(absURL)
					}
					normAbsURL, err := normalizeURL(absURL.String())
					if err != nil {
						continue
					}

					localPath, ok := p.urlToLocal[normAbsURL]
					if ok && localPath != "" {
						relPath, err := filepath.Rel(filepath.Dir(htmlFilename), localPath)
						if err == nil {
							n.Attr[i].Val = filepath.ToSlash(relPath)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	out, err := os.Create(htmlFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Println(err)
		}
	}()

	return html.Render(out, doc)
}
