package email

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/abhissng/neuron/utils/helpers"
	gomail "gopkg.in/mail.v2"
)

// fetchAttachment fetches attachment data from a URL or local file
func fetchAttachment(att string) ([]byte, string, string, error) {
	if helpers.IsURL(att) {
		parsed, err := url.Parse(att)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to parse URL %s: %w", att, err)
		}
		resp, err := http.Get(parsed.String())
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to GET URL %s: %w", att, err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		if resp.StatusCode != http.StatusOK {
			return nil, "", "", fmt.Errorf("bad status fetching URL %s: %d", att, resp.StatusCode)
		}
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, resp.Body); err != nil {
			return nil, "", "", fmt.Errorf("error reading body from URL %s: %w", att, err)
		}
		fname := filepath.Base(resp.Request.URL.Path)
		if fname == "" {
			fname = "attachment"
		}
		ext := path.Ext(fname)
		ctype := mime.TypeByExtension(ext)
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		return buf.Bytes(), fname, ctype, nil
	}

	// Local file
	data, err := io.ReadAll(mustOpen(att))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to read local file %s: %w", att, err)
	}
	fname := filepath.Base(att)
	ext := path.Ext(fname)
	ctype := mime.TypeByExtension(ext)
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	return data, fname, ctype, nil
}

// mustOpen opens a file and panics on error (used internally)
func mustOpen(path string) io.Reader {
	f, err := os.Open(path)
	if err != nil {
		return bytes.NewReader(nil)
	}
	return f
}

// attachFiles attaches files to the email message
func attachFiles(m *gomail.Message, attachments []string) error {
	for _, att := range attachments {
		if helpers.IsURL(att) {
			// fetch
			parsed, err := url.Parse(att)
			if err != nil {
				return fmt.Errorf("failed to parse URL %s: %w", att, err)
			}
			resp, err := http.Get(parsed.String())
			if err != nil {
				return fmt.Errorf("failed to GET URL %s: %w", att, err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("bad status fetching URL %s: %d", att, resp.StatusCode)
			}
			// read all
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, resp.Body); err != nil {
				return fmt.Errorf("error reading body from URL %s: %w", att, err)
			}
			// derive filename
			fname := filepath.Base(resp.Request.URL.Path)
			if fname == "" {
				fname = "attachment"
			}
			// content type
			ext := path.Ext(fname)
			ctype := mime.TypeByExtension(ext)
			if ctype == "" {
				ctype = "application/octet-stream"
			}
			// AttachReader
			m.AttachReader(fname, buf, gomail.SetHeader(map[string][]string{
				"Content-Type":        {ctype + `; name="` + fname + `"`},
				"Content-Disposition": {`attachment; filename="` + fname + `"`},
			}))
		} else {
			// local file
			m.Attach(att)
		}
	}
	return nil
}
