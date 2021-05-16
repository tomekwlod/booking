package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

// func ChainReccur(h http.Handler, m ...Middleware) http.Handler {
// 	// if our chain is done, use the final handler
// 	if len(m) == 0 {
// 		return h
// 	}

// 	// otherwise nest the handle
// 	return m[0](ChainReccur(h, m[1:cap(m)-1]...))
// }

////https://gist.github.com/enricofoltran/10b4a980cd07cb02836f70a4ab3e72d7
// logger := log.New(os.Stdout, "http: ", log.LstdFlags)
// logger.Println("Server is starting...")
// func logging(logger *log.Logger) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			defer func() {
// 				requestID, ok := r.Context().Value(requestIDKey).(string)
// 				if !ok {
// 					requestID = "unknown"
// 				}
// 				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
// 			}()
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := extendedResponseWriter{ResponseWriter: w}

		// Call the next middleware/handler in chain
		h.ServeHTTP(&rw, r)

		// defer func() {

		requestID, ok := r.Context().Value(requestIDKey).(string)
		if !ok {
			requestID = "unknown"
		}

		log.Printf("[%s] %s %s %d %v %dbytes {%s} [%s]",
			requestGetRemoteAddress(r),
			r.Method,
			r.URL.String(),
			rw.status,
			time.Since(start),
			int64(rw.length),
			r.Header.Get("User-Agent"),
			requestID,
			//r.Header.Get("Referer"),
		)
		// }()

	})
}

// test first and add to the regular middlewares
type key int

const (
	requestIDKey key = 0
)

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func supportXHTTPMethodOverrideMiddleware() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := r.Header.Get("X-HTTP-Method-Override")
			if len(m) > 0 {
				r.Method = m
			}
			h.ServeHTTP(w, r)
		})
	}
}
func withHeaderMiddleware(key, value string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(key, value)
			h.ServeHTTP(w, r)
		})
	}
}

func recoverMiddleware() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			defer func() {
				if err := recover(); err != nil {

					log.Printf("[recover mode] An error occuren: %+v", err)

					writeError(w, errInternalServer)

					return
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}

func allowCorsMiddleware(w http.ResponseWriter, r *http.Request) {
	// if r.Method == "OPTIONS" {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, PUT")

	// w.WriteHeader(200)
	// }
}

// https://stackoverflow.com/a/24818638/1800372
// func EnableCORS() Adapter {
//     return func(h http.Handler) http.Handler {
//         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

//             if origin := r.Header.Get("Origin"); origin != "" {
//                 w.Header().Set("Access-Control-Allow-Origin", origin)
//                 w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
//                 w.Header().Set("Access-Control-Allow-Headers",
//                     "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
//             }
//             // Stop here if its Preflighted OPTIONS request
//             if r.Method == "OPTIONS" {
//                 return
//             }
//             h.ServeHTTP(w, r)
//         })
//     }
// }
