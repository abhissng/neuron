package email

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"

	"github.com/abhissng/neuron/utils/helpers"
	gomail "gopkg.in/mail.v2"
)

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
			defer resp.Body.Close()
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
