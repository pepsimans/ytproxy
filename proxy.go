package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	url  = "https://www.youtube.com/watch?v=%s"
	ytdl = "./youtube-dl"
)

const (
	methodGetFormat = "GET_FORMAT"
	defaultDpi	  = "720p"
	defaultFormat = "best[ext=mp4]"

	httpCodeWrongRequest = 400
	httpCodeError = 500
)

const (
	clearCachePeriod = 2 * time.Minute
	cacheTTL = 1 * time.Hour

	socketTimeOut = 10 * time.Minute
)

// Entry declares cache record
type Entry struct {
	url []byte
	expire time.Time
}

var (
	regex, _ = regexp.Compile(`(\d+)\s+mp4\s*(\S*)\s*([^12][0-9]*p).+Hz`)
	cache = make(map[string]Entry, 100)

	cacheMutex = &sync.RWMutex{}
)

func main() {
	router := mux.NewRouter()

	ticker := time.NewTicker(clearCachePeriod)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C

			log.Print("Start cache clean")
			now := time.Now()

			for id, item := range cache {
				if item.expire.Before(now) {
					cacheMutex.Lock()
					delete(cache, id)

					log.Print("Delete item " + id)

					cacheMutex.Unlock()
				}
			}

			time.Sleep(100 * time.Microsecond)
		}
	}()


	router.HandleFunc("/get_url", func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("Request from IP: %v", request.RemoteAddr)

		q := request.URL.Query()
		id := q.Get("id")

		cacheMutex.Lock()
		data, ok := cache[id]
		cacheMutex.Unlock()

		if ok {
			log.Print("Cache hit")
			_ , _ = writer.Write(data.url)

			return
		}

		ctx := request.Context()
		f := q.Get("f")

		log.Print("Sent-" + request.Header.Get("X-Auth"))
		log.Print("Set - " + os.Getenv("SERIAL_NO"))

		if id == "" || os.Getenv("SERIAL_NO") != request.Header.Get("X-Auth") {
			writer.WriteHeader(httpCodeWrongRequest)
			return
		}

		if f == "" {
			f = defaultDpi
		}

		out, err := proxyCall(ctx, id, methodGetFormat)
		if err != nil {
			writer.WriteHeader(httpCodeError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		// Ok, now we need to find best format
		res := regex.FindAllSubmatch(out, -1)

		l := len(res)
		selected := defaultFormat

		for i := range res {
			k := l - (i + 1)
			dpi := string(res[k][3])
			fid := string(res[k][1])

			if f == dpi || selected == defaultFormat {
				selected = fid
			}
		}

		log.Print("Selected format: " + selected)

		out2, err := proxyCall(ctx, id, selected)
		if err != nil {
			writer.WriteHeader(httpCodeError)
			_, _ = writer.Write([]byte(err.Error()))

			return
		}

		cacheMutex.Lock()
		cache[id] = Entry{url: out2, expire: time.Now().Add(cacheTTL)}
		cacheMutex.Unlock()

		log.Print(string(out2))
		_, _ = writer.Write(out2)
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         ":5080",
		WriteTimeout: socketTimeOut,
		ReadTimeout:  socketTimeOut,
	}

	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(closed)
	}()

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Error during web server start %v", err)
	}

	log.Println("Web server started")

	<-closed
}

// proxyCall Make call to youtube-dl
func proxyCall(ctx context.Context, id, opt string) ([]byte, error) {
	var args []string

	youtubeURL := fmt.Sprintf(url, id)

	if opt == methodGetFormat {
		args = []string{"--list-formats", youtubeURL}
	} else {
		args = []string{"--format=" + opt, "-g", youtubeURL}
	}

	return exec.CommandContext(ctx, ytdl, args...).Output()
}
