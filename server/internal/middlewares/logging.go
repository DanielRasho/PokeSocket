package middlewares

// import (
// 	"net/http"
// 	"time"

// 	"github.com/aws/aws-lambda-go/events"
// 	"github.com/rs/zerolog/log"
// )

// // Logs detailed information of an exception on the API when called.
// // Parameters:
// //   - err : Error that caused the exception
// //   - r : ApiGateway that handled the routing
// //   - status : The response message that the API is planning to return
// //   - body : Initial JSON payload that was sent by the client
// func LogErrorRequest(err error, r *events.APIGatewayProxyRequest, statusCode int, body []byte, response []byte) {

// 	event := log.Error().Err(err).
// 		Str("method", r.HTTPMethod).
// 		Str("route", r.Path).
// 		Int("status", statusCode)

// 	if len(body) > 0 {

// 		event.RawJSON("body", body)
// 	}
// 	if len(response) > 0 {
// 		event.RawJSON("response", response)
// 	}
// 	event.Msg("Request failed")
// }

// // USED ONLY ON DEVELOPMENT BUILD
// // Logging logs the details of each incoming HTTP request.
// func Logging(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Wrap the ResponseWriter to capture the status code
// 		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

// 		start := time.Now()
// 		next.ServeHTTP(wrappedWriter, r)
// 		duration := time.Since(start)

// 		// Log the request details using zerolog
// 		log.Debug().
// 			Str("method", r.Method).
// 			Str("route", r.URL.Path).
// 			Int("status", wrappedWriter.statusCode).
// 			Int64("unix", start.Unix()).
// 			Dur("duration", duration).
// 			Msg("Request processed")
// 	})
// }

// // responseWriter is a wrapper around http.ResponseWriter to capture the status code
// type responseWriter struct {
// 	http.ResponseWriter
// 	statusCode  int
// 	wroteHeader bool // Prevent multiple calls to WriteHeader
// }

// // WriteHeader captures the status code and ensures it's only called once
// func (rw *responseWriter) WriteHeader(code int) {
// 	if rw.wroteHeader {
// 		return // Prevent superfluous calls
// 	}
// 	rw.wroteHeader = true
// 	rw.statusCode = code
// 	rw.ResponseWriter.WriteHeader(code)
// }

// // Write ensures the header is written before writing the body
// func (rw *responseWriter) Write(b []byte) (int, error) {
// 	if !rw.wroteHeader {
// 		rw.WriteHeader(http.StatusOK) // Default to 200 if WriteHeader wasn't called
// 	}
// 	return rw.ResponseWriter.Write(b)
// }
