package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	nats "github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel"
	"github.com/nsyszr/lcm/pkg/storage/memory"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type deviceControlServer struct {
	quitCh chan bool
	doneCh chan bool

	nc    *nats.Conn
	errCh chan error
	wg    sync.WaitGroup
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func newDeviceControlServer() (*deviceControlServer, error) {
	s := &deviceControlServer{
		// mgr:     memory.NewMemoryManager(),
		quitCh: make(chan bool),
		doneCh: make(chan bool),

		errCh: make(chan error, 1),
		wg:    sync.WaitGroup{},
	}

	nc, err := nats.Connect(nats.DefaultURL,
		nats.DrainTimeout(10*time.Second),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			fmt.Printf("\n\nerror handler: %s\n\n", err)
			s.errCh <- err
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			fmt.Printf("\n\nclosed handler\n\n")
			s.wg.Done()
		}),
		nats.DisconnectHandler(func(_ *nats.Conn) {
			// TODO(DGL) this method is called twice when NATS server is going
			// offline. 1st when server gone and 2nd when the shutdown/drain is
			// initiated.
			fmt.Printf("\n\ndisconnect handler\n\n")
			// s.wg.Done()
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			//s.quitCh <- os.Interrupt
		}))
	if err != nil {
		return nil, err
	}

	s.nc = nc

	return s, nil
}

func (s *deviceControlServer) Serve() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(logger())
	// e.HTTPErrorHandler = errorx.JSONErrorHandler

	// Create the controller
	ctrl := controlchannel.NewController(s.nc, memory.NewStore())
	ctrl.Subscribe()

	// Register API endpoints
	deviceControlHandler := devicecontrol.NewHandler(ctrl)
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
	<-s.quitCh
	log.Info("Shutdown signal received")

	// Create a 10 second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the echo web server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Error(err)
	}

	// We've done!
	s.doneCh <- true
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

func (s *deviceControlServer) Shutdown() {
	if s.nc != nil {
		s.nc.Drain()
	}

	// Send the quit signal to the server.ServeAPI() routine
	s.quitCh <- true

	// Wait up to 10 seconds
	select {
	case <-s.doneCh:
		log.Info("Shutdown server successful")
	case <-time.After(10 * time.Second):
		log.Error("Shutdown server failed")
	}
}

func RunServeDeviceControl() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		s, err := newDeviceControlServer()
		if err != nil {
			log.Error("failed to create new server instance: ", err)
			os.Exit(1)
		}

		go s.Serve()

		// Wait for interrupt signal to gracefully shutdown the server
		quitCh := make(chan os.Signal)
		signal.Notify(quitCh, os.Interrupt)
		<-quitCh

		// Shutdown the server
		s.Shutdown()
	}
}
