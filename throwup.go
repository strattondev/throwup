package throwup

import (
	"github.com/google/uuid"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var templates = template.Must(template.ParseFiles("templates/upload.html"))

type upload struct {
	Url string
}

type ThrowUp interface {
	Run(mux *http.ServeMux)
}

func New(up Config) ThrowUp {
	cleanConfig(&up)

	return throwUp{up, make(map[string]bool)}
}

func cleanConfig(up *Config) {
	if up.EndpointCount <= 0 {
		up.EndpointCount = 1
	}

	if len(up.StorageBase) == 0 {
		up.StorageBase = "/tmp"
	}

	if len(up.PostSuccessMessage) == 0 {
		up.PostSuccessMessage = "Thank you for you upload"
	}

	if len(up.PostSingleUseMessage) == 0 {
		up.PostSingleUseMessage = "Sorry, you can no longer upload files"
	}

	up.StorageBase = filepath.Clean(up.StorageBase)
}

type Config struct {
	EndpointCount        int
	StorageBase          string
	PostSuccessMessage   string
	PostSingleUse        bool
	PostSingleUseMessage string
}

type throwUp struct {
	ThrowUpConfig Config
	endpoints     map[string]bool
}

func (throwUp throwUp) Run(mux *http.ServeMux) {
	for i := 0; i < throwUp.ThrowUpConfig.EndpointCount; i++ {
		u := "/" + uuid.New().String()
		mux.HandleFunc(u, throwUp.uploadHandler)
		log.Printf(u)
	}
}

func (throwUp throwUp) display(w http.ResponseWriter, template string, data interface{}) {
	templates.ExecuteTemplate(w, template+".html", data)
}

func (throwUp throwUp) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if throwUp.ThrowUpConfig.PostSingleUse && throwUp.endpoints[r.RequestURI] {
		w.Write([]byte(throwUp.ThrowUpConfig.PostSingleUseMessage))
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

			dst, err := os.Create(throwUp.ThrowUpConfig.StorageBase + "/" + part.FileName())

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

		w.Write([]byte(throwUp.ThrowUpConfig.PostSuccessMessage))
		throwUp.endpoints[r.RequestURI] = true
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
