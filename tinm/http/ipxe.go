package http

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/homeric-io/tinm/tinm/storage/storagepb"
	"errors"
)

const ipxeBootstrap = `#!ipxe
chain ipxe?uuid=${uuid}&mac=${mac:hexhyp}&domain=${domain}&hostname=${hostname}&serial=${serial}&arch=${buildarch:uristring}
`

var ipxeBootTemplate = template.Must(template.New("iPXE config").Parse(`#!ipxe
kernel {{.Kernel}}{{range $arg := .Args}} {{$arg}}{{end}}
{{- range $element := .Initrd }}
initrd {{$element}}
{{- end}}
boot
`))

var ipxeChainTemplate = template.Must(template.New("iPXE config").Parse(`#!ipxe
chain {{ . }}
`))

// ipxeInspect returns a handler that responds with the iPXE script to gather
// client machine data and chainload to the ipxeHandler.
func ipxeInspect() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, ipxeBootstrap)
	}
	return http.HandlerFunc(fn)
}

// ipxeBoot returns a handler which renders the iPXE boot script for the
// requester.
func (s *Server) ipxeHandler() http.Handler {
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
		}).Debug("Matched an iPXE config")

		var buf bytes.Buffer
		switch v := profile.GetBootMode().(type) {
			case *storagepb.Profile_Boot:
				err = ipxeBootTemplate.Execute(&buf, v.Boot)
			case *storagepb.Profile_Chain:
				err = ipxeChainTemplate.Execute(&buf, v.Chain)
			default:
				err = errors.New("no boot or chain mode available")
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
