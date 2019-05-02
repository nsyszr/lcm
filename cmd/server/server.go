package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nsyszr/lcm/pkg/devicecontrol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type server struct {
	quitAPI chan bool
	doneAPI chan bool
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func newServer() (*server, error) {
	return &server{
		// mgr:     memory.NewMemoryManager(),
		quitAPI: make(chan bool),
		doneAPI: make(chan bool),
	}, nil
}

func (s *server) ServeAPI() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(logger())
	// e.HTTPErrorHandler = errorx.JSONErrorHandler

	// Register API endpoints
	deviceControlHandler := devicecontrol.NewHandler()
	deviceControlHandler.RegisterRoutes(e)

	// Register devicecontrol endpoint
	// e.Any("/devicecontrol/v1", devicecontrol.Handler())

	go func() {
		log.WithFields(log.Fields{
			"host": "",
			"port": 4001,
		}).Info("Starting server")

		if err := e.Start(fmt.Sprintf("%s:%d", "", 4001)); err != nil {
			e.Logger.Info("Shutting down the server")
		}
	}()

	// Wait until receiving the quit signal
	<-s.quitAPI
	log.Info("Shutdown signal received")

	// Create a 10 second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the echo web server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Error(err)
	}

	// We've done!
	s.doneAPI <- true
}

// Logger returns a middleware that logs HTTP requests.
func logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			reqSizeStr := req.Header.Get(echo.HeaderContentLength)
			if reqSizeStr == "" {
				reqSizeStr = "0"
			}
			reqSize, err := strconv.ParseInt(reqSizeStr, 10, 0)
			if err != nil {
				reqSize = -1
			}
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			log.WithFields(log.Fields{
				"timestamp":     stop.Format(time.RFC3339),
				"id":            id,
				"remote_ip":     c.RealIP(),
				"host":          req.Host,
				"method":        req.Method,
				"uri":           req.RequestURI,
				"protocol":      req.Proto,
				"user_agent":    req.UserAgent(),
				"status":        res.Status,
				"status_text":   http.StatusText(res.Status),
				"referer":       req.Referer(),
				"error":         errMsg,
				"bytes_in":      reqSize,
				"bytes_out":     res.Size,
				"latency":       stop.Sub(start).Nanoseconds(),
				"latency_human": stop.Sub(start).String(),
			}).Infof("%s %s %s %d %s", req.Method, req.RequestURI, req.Proto,
				res.Status, strconv.FormatInt(res.Size, 10))

			return err
		}
	}
}

func (s *server) ShutdownAPI() {
	// Send the quit signal to the server.ServeAPI() routine
	s.quitAPI <- true

	// Wait up to 10 seconds
	select {
	case <-s.doneAPI:
		log.Info("Shutdown server successful")
	case <-time.After(10 * time.Second):
		log.Error("Shutdown server failed")
	}
}

func RunServe() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		s, err := newServer()
		if err != nil {
			log.Error("failed to create new server instance: ", err)
			os.Exit(1)
		}

		go s.ServeAPI()

		// Wait for interrupt signal to gracefully shutdown the server
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit

		// Shutdown the server
		s.ShutdownAPI()
	}
}
