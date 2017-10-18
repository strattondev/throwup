package throwup

import (
	"net/http"
	"io"
	"os"
	"github.com/google/uuid"
	"log"
	"html/template"
	"path/filepath"
)

var templates = template.Must(template.ParseFiles("templates/upload.html"))

type upload struct {
	Url string
}

type ThrowUp interface {
	Run(mux *http.ServeMux)
}

func New(up ThrowUpConfig) ThrowUp {
	cleanConfig(&up)

	return throwUp{up, make(map[string]bool)}
}

func cleanConfig(up *ThrowUpConfig) {
	if up.endpointCount <= 0 {
		up.endpointCount = 1
	}

	if len(up.storageBase) == 0 {
		up.storageBase = "/tmp"
	}

	if len(up.postSuccessMessage) == 0 {
		up.postSuccessMessage = "Thank you for you upload"
	}

	if len(up.postSingleUseMessage) == 0 {
		up.postSingleUseMessage = "Sorry, you can no longer upload files"
	}

	up.storageBase = filepath.Clean(up.storageBase)
}

type ThrowUpConfig struct {
	endpointCount int
	storageBase   string
	postSuccessMessage string
	postSingleUse bool
	postSingleUseMessage string
}

type throwUp struct {
	ThrowUpConfig ThrowUpConfig
	endpoints map[string]bool
}

func (throwUp throwUp) Run(mux *http.ServeMux) {
	for i := 0; i < throwUp.ThrowUpConfig.endpointCount; i++ {
		u := "/" + uuid.New().String()
		mux.HandleFunc(u, throwUp.uploadHandler)
		log.Printf(u)
	}
}

func (throwUp throwUp) display(w http.ResponseWriter, template string, data interface{}) {
	templates.ExecuteTemplate(w, template + ".html", data)
}

func (throwUp throwUp) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if throwUp.ThrowUpConfig.postSingleUse && throwUp.endpoints[r.RequestURI] {
		w.Write([]byte(throwUp.ThrowUpConfig.postSingleUseMessage))
		return
	}

	switch r.Method {
	case "GET":
		throwUp.display(w, "upload", upload{r.RequestURI})
	case "POST":
		reader, err := r.MultipartReader()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			if part.FileName() == "" {
				continue
			}

			dst, err := os.Create(throwUp.ThrowUpConfig.storageBase + "/" + part.FileName())

			defer dst.Close()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Write([]byte(throwUp.ThrowUpConfig.postSuccessMessage))
		throwUp.endpoints[r.RequestURI] = true
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}