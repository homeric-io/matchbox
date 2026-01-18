package http

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/homeric-io/tinm/tinm/storage/storagepb"
	"errors"
)

var grubTemplate = template.Must(template.New("GRUB2 config").Parse(`default=0
fallback=1
timeout=1
menuentry "CoreOS (EFI)" {
echo "Loading kernel"
linuxefi "{{.Kernel}}"{{range $arg := .Args}} {{$arg}}{{end}}
echo "Loading initrd"
initrdefi {{ range $element := .Initrd }} "{{$element}}"{{end}}
}
menuentry "CoreOS (BIOS)" {
echo "Loading kernel"
linux "{{.Kernel}}"{{range $arg := .Args}} {{$arg}}{{end}}
echo "Loading initrd"
initrd {{ range $element := .Initrd }} "{{$element}}"{{end}}
}
`))

// grubHandler returns a handler which renders a GRUB2 config for the
// requester.
func (s *Server) grubHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		profile, err := profileFromContext(ctx)
		if err != nil {
			s.logger.WithFields(logrus.Fields{
				"labels": labelsFromRequest(nil, req),
			}).Infof("No matching profile")
			http.NotFound(w, req)
			return
		}

		// match was successful
		s.logger.WithFields(logrus.Fields{
			"labels":  labelsFromRequest(nil, req),
			"profile": profile.Id,
		}).Debug("Matched a GRUB config")

		var buf bytes.Buffer
		switch v := profile.GetBootMode().(type) {
			case *storagepb.Profile_Boot:
				err = grubTemplate.Execute(&buf, v.Boot)
			case *storagepb.Profile_Chain:
				err = errors.New("Chain boot mode not supported for grub http api")
			}
		if err != nil {
			s.logger.Errorf("error rendering template: %v", err)
			http.NotFound(w, req)
			return
		}
		if _, err := buf.WriteTo(w); err != nil {
			s.logger.Errorf("error writing to response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	return http.HandlerFunc(fn)
}
