package webapp

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"text/template"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "App: ", log.Ldate|log.Ltime|log.Lmsgprefix)
}

type App struct {
	server *http.Server

	mutex              sync.RWMutex
	recentStatusUpdate *StatusUpdate

	updateChannel chan *StatusUpdate
}

func NewApp(listenAddr string) *App {
	app := App{}
	app.updateChannel = make(chan *StatusUpdate)

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}
	tmpl := template.Must(template.New("tmpl").Funcs(funcMap).Parse(
		`OpenVPN active connections as of {{ .Time }}:

{{ range $idx, $item := .Lines -}}
{{ inc $idx }}. Common Name: {{ $item.CommonName}}
 Real Address: {{ $item.RealAddress}} Virtual Address: {{ $item.VirtualAddress }}, Virtual IPv6 Address: {{ $item.VirtualIPv6Address }}
 Bytes Received: {{ $item.BytesReceived }} Bytes Sent: {{ $item.BytesSent }}
 Connected Since: {{ $item.ConnectedSince }}
 Username: {{ $item.Username }} Client ID: {{ $item.ClientId }} Peer ID: {{ $item.PeerId }} Data Channel Cipher: {{ $item.DataChannelCipher }}

{{ else -}}
No active connections
{{ end -}}`))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		app.mutex.RLock()
		update := app.recentStatusUpdate
		app.mutex.RUnlock()

		if update == nil {
			fmt.Fprint(resp, "No data")
			return
		}

		if update.Error != nil {
			fmt.Fprintf(resp, "%s", update.Error)
		} else {
			tmpl.Execute(resp, update.Status)
		}
	})
	app.server = &http.Server{Addr: listenAddr, Handler: mux}

	return &app
}

func (app *App) UpdateChannel() chan<- *StatusUpdate {
	return app.updateChannel
}

func (app *App) Run() error {
	go func() {
		for update := range app.updateChannel {
			if update.Error == nil && update.Status == nil {
				logger.Println("empty update status")
			} else {
				app.mutex.Lock()
				app.recentStatusUpdate = update
				app.mutex.Unlock()
			}
		}
	}()

	logger.Printf("listening for HTTP connections on %s", app.server.Addr)

	return app.server.ListenAndServe()
}
