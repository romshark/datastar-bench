package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/romshark/datastar-bench/template"

	"github.com/a-h/templ"
	"github.com/andybalholm/brotli"
	"github.com/starfederation/datastar-go/datastar"
)

//go:generate go run github.com/a-h/templ/cmd/templ@latest generate

func main() {
	fHost := flag.String("host", ":8080", "host address")
	flag.Parse()

	mux := http.NewServeMux()

	handle := func(pattern string, handlerFn http.HandlerFunc) {
		mux.Handle(pattern, withBrotli(handlerFn))
	}

	waitAckFatMorph := make(chan struct{})

	handle("/", getIndex)
	handle("/inputs/{$}", getInputs)
	handle("/checkboxes/{$}", getCheckboxes)
	handle("/textbind/{$}", getTextbind)
	handle("/textexpr/{$}", getTextexpr)
	handle("/condshow/{$}", getCondshow)
	handle("/ssepatchrep/{$}", getSSEPatchRep)
	handle("/ssepatchmorph/{$}", getSSEPatchMorph)
	handle("/ssepatchfatmorph/{$}", makeGetSSEPatchFatMorph(waitAckFatMorph))
	handle("POST /ssepatchfatmorph/ack/{$}", makeGetSSEPatchFatMorphAck(waitAckFatMorph))

	slog.Info("listeninig", slog.String("host", *fHost))
	panic(http.ListenAndServe(*fHost, mux))
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}
	render(w, r, template.PageIndex(), "page index")
}

func getInputs(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		Num float64 `json:"num"`
	}
	start := time.Now()
	renderPage(w, r, "page inputs", template.PageInputs(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			patch(sse,
				template.FragUpdate(
					time.Since(start),
					template.FragInputsContent(int(s.Num)),
				),
				"inputs update")
			return nil
		},
	)
}

func getCheckboxes(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		Num float64 `json:"num"`
	}
	start := time.Now()
	renderPage(w, r, "page checkboxes", template.PageCheckboxes(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			patch(sse,
				template.FragUpdate(
					time.Since(start),
					template.FragCheckboxesContent(int(s.Num)),
				),
				"checkboxes update")
			return nil
		},
	)
}

func getTextbind(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		Num float64 `json:"num"`
	}
	start := time.Now()
	renderPage(w, r, "page textbind", template.PageTextBind(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			patch(sse,
				template.FragUpdate(
					time.Since(start),
					template.FragTextBindContent(int(s.Num)),
				),
				"text bind update")
			return nil
		},
	)
}

func getTextexpr(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		Num float64 `json:"num"`
	}
	start := time.Now()
	renderPage(w, r, "page textexpr", template.PageTextExpr(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			patch(sse,
				template.FragUpdate(
					time.Since(start),
					template.FragTextExprContent(int(s.Num)),
				),
				"text expr update")
			return nil
		},
	)
}

func getCondshow(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		NumRed  float64 `json:"numred"`
		NumBlue float64 `json:"numblue"`
	}
	start := time.Now()
	renderPage(w, r, "page condshow", template.PageCondShow(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			patch(sse,
				template.FragUpdate(
					time.Since(start),
					template.FragCondShowContent(int(s.NumRed), int(s.NumBlue)),
				),
				"cond show update")
			return nil
		},
	)
}

var patchOptsReplace = []datastar.PatchElementOption{datastar.WithModeReplace()}

func getSSEPatchRep(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		RatePerSec float64 `json:"ratepersec"`
	}
	start := time.Now()
	renderPage(w, r, "page sse patch replace", template.PageSSEPatchRep(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			interval := time.Second / time.Duration(s.RatePerSec)
			tk := time.NewTicker(interval)
			initialStart := time.Now()
		LOOP:
			for c := int64(0); ; {
				select {
				case <-sse.Context().Done():
					break LOOP
				case <-tk.C:
					start = time.Now()
					c++
					patch(sse,
						template.FragUpdate(
							time.Since(start),
							template.FragSSEPatchRepContent(c, initialStart, time.Now()),
						),
						"ssepatch replace update",
						patchOptsReplace...)
				}
			}
			return nil
		},
	)
}

func getSSEPatchMorph(w http.ResponseWriter, r *http.Request) {
	type Sigs struct {
		RateLimit  bool    `json:"ratelimit"`
		RatePerSec float64 `json:"ratepersec"`
	}
	start := time.Now()
	renderPage(w, r, "page sse patch morph", template.PageSSEPatchMorph(),
		func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
			initialStart := time.Now()
			if s.RateLimit {
				interval := time.Second / time.Duration(s.RatePerSec)
				tk := time.NewTicker(interval)
			LOOP:
				for c := int64(0); ; {
					select {
					case <-sse.Context().Done():
						break LOOP
					case <-tk.C:
						start = time.Now()
						c++
						patch(sse,
							template.FragUpdate(
								time.Since(start),
								template.FragSSEPatchMorphContent(
									c, initialStart, time.Now(),
								),
							),
							"ssepatch morph update")
					}
				}
			} else {
				// Unlimited rate, shoot as fast as you can 🚀
				for c := int64(0); sse.Context().Err() == nil; {
					start = time.Now()
					c++
					patch(sse,
						template.FragUpdate(
							time.Since(start),
							template.FragSSEPatchMorphContent(
								c, initialStart, time.Now(),
							),
						),
						"ssepatch morph update")
				}
			}
			return nil
		},
	)
}

var dataFatMorph1 = func() (data *template.DataFatMorph) {
	fc, err := os.ReadFile("./chat-resp-1.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(fc, &data)
	if err != nil {
		panic(err)
	}
	return data
}()

var dataFatMorph2 = func() (data *template.DataFatMorph) {
	fc, err := os.ReadFile("./chat-resp-2.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(fc, &data)
	if err != nil {
		panic(err)
	}
	return data
}()

func makeGetSSEPatchFatMorph(waitAck <-chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Sigs struct {
			Run        bool    `json:"run"`
			RateLimit  bool    `json:"ratelimit"`
			RatePerSec float64 `json:"ratepersec"`
		}
		start := time.Now()

		renderPage(w, r, "page sse patch fat morph", template.PageSSEPatchFatMorph(),
			func(sse *datastar.ServerSentEventGenerator, s Sigs) error {
				patch := func(counter int64, initialStart time.Time) {
					data := dataFatMorph1
					if counter%2 == 0 {
						data = dataFatMorph2
					}
					patch(sse,
						template.FragUpdate(
							time.Since(start),
							template.FragSSEPatchFatMorphContent(
								counter, initialStart, time.Now(), data,
							),
						), "ssepatch fat morph update")
				}

				initialStart := time.Now()
				if !s.Run {
					patch(0, initialStart)
					return nil
				}
				if s.RateLimit {
					interval := time.Second / time.Duration(s.RatePerSec)
					tk := time.NewTicker(interval)
				LOOP:
					for c := int64(0); ; {
						select {
						case <-sse.Context().Done():
							break LOOP
						case <-tk.C:
							start = time.Now()
							c++
							patch(c, initialStart)
							select { // Wait until stream closes or ack is received.
							case <-sse.Context().Done():
							case <-waitAck:
							}
						}
					}
				} else {
					// Unlimited rate, shoot as fast as you can 🚀
					for c := int64(0); sse.Context().Err() == nil; {
						start = time.Now()
						c++
						patch(c, initialStart)
						select { // Wait until stream closes or ack is received.
						case <-sse.Context().Done():
						case <-waitAck:
						}
					}
				}
				return nil
			},
		)
	}
}

func makeGetSSEPatchFatMorphAck(waitAck chan<- struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		waitAck <- struct{}{}
	}
}

var defaultSSEOpts = []datastar.SSEOption{
	datastar.WithCompression(datastar.WithBrotli(datastar.WithBrotliLGWin(15))),
}

func renderPage[Signals any](
	w http.ResponseWriter, r *http.Request,
	compName string,
	page templ.Component,
	handleSSE func(sse *datastar.ServerSentEventGenerator, s Signals) error,
) {
	if !isReqDS(r) {
		render(w, r, page, compName)
		return
	}

	sse := datastar.NewSSE(w, r, defaultSSEOpts...)

	sig, ok := readSignals[Signals](w, r)
	if !ok {
		return
	}
	if err := sse.PatchElementTempl(page); err != nil {
		const code = http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
	}
	if err := handleSSE(sse, sig); err != nil {
		const code = http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
	}
}

func render(
	w http.ResponseWriter, r *http.Request, comp templ.Component, compName string,
) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := comp.Render(r.Context(), w); err != nil {
		slog.Info("rendering", slog.String("component", compName), slog.Any("err", err))
	}
}

func patch(
	sse *datastar.ServerSentEventGenerator, comp templ.Component, compName string,
	opts ...datastar.PatchElementOption,
) (ok bool) {
	if err := sse.PatchElementTempl(comp, opts...); err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Info("patching",
				slog.String("component", compName), slog.Any("err", err))
		}
		return false
	}
	return true
}

func readSignals[S any](
	w http.ResponseWriter, r *http.Request,
) (signals S, ok bool) {
	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, "bad signals: "+err.Error(), http.StatusBadRequest)
		return signals, false
	}
	return signals, true
}

func isReqDS(r *http.Request) bool { return r.Header.Get("Datastar-Request") == "true" }

type brResponseWriter struct {
	http.ResponseWriter
	bw *brotli.Writer
}

func (w *brResponseWriter) Write(b []byte) (int, error) { return w.bw.Write(b) }

func withBrotli(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't compress SSE streams, let DS SDK handle it.
		if strings.Contains(r.Header.Get("Accept"), "text/event-stream") ||
			r.Header.Get("Datastar-Request") == "true" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "br")
		bw := brotli.NewWriterLevel(w, brotli.BestCompression)
		defer func() { _ = bw.Close() }()

		brw := &brResponseWriter{ResponseWriter: w, bw: bw}
		next.ServeHTTP(brw, r)
	})
}
