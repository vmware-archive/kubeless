package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	stdinChannel  = 0
	stdoutChannel = 1
	stderrChannel = 2
	errChannel    = 3
)

// Cmd stores information relevant to an individual remote command being run
type Cmd struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// RoundTripCallback is suitable to use with `ExecRoundTripper` and will
// copy data to/from stdio channels.  The returned `Response` is
// currently always `nil`.
func (c *Cmd) RoundTripCallback(conn *websocket.Conn) (*http.Response, error) {
	errChan := make(chan error, 3)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if c.Stdin == nil {
			return
		}
		buf := make([]byte, 1025) // NB: first byte is fixed
		buf[0] = stdinChannel
		for {
			n, err := c.Stdin.Read(buf[1:])
			err2 := websocket.Message.Send(conn, buf[:n+1])
			if err == nil && err2 != nil {
				err = err2
			}
			if err == io.EOF {
				break
			} else if err != nil {
				errChan <- err
				return
			}
		}
		const closeStatusNormal = 1000
		conn.WriteClose(closeStatusNormal)
	}()
	go func() {
		defer wg.Done()
		for {
			var buf []byte
			err := websocket.Message.Receive(conn, &buf)
			if err == io.EOF {
				break
			} else if err != nil {
				errChan <- err
				return
			}
			if len(buf) == 0 {
				logrus.Debug("Received empty message, skipping")
				continue
			}
			logrus.Debugf("Received %dB message for channel %d", len(buf)-1, buf[0])
			var w io.Writer
			switch buf[0] {
			case stdoutChannel:
				w = c.Stdout
			case stderrChannel:
				w = c.Stderr
			case errChannel:
				errChan <- fmt.Errorf("Error from remote command: %s", buf[1:])
				return
			default:
				logrus.Infof("Ignoring message for unknown channel %d", buf[0])
				continue
			}
			if w == nil {
				logrus.Infof("Ignoring message for nil channel %d", buf[0])
				continue
			}
			_, err = w.Write(buf[1:])
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	wg.Wait()
	close(errChan)
	err := <-errChan
	return &http.Response{
		Status:     "OK",
		StatusCode: 200,
	}, err
}

// A RoundTripCallback is used to process the websocket from an
// individual command execution.
type RoundTripCallback func(conn *websocket.Conn) (*http.Response, error)

// WebsocketRoundTripper is an http.RoundTripper that invokes a
// callback on a websocket connection.
type WebsocketRoundTripper struct {
	TLSConfig *tls.Config
	Do        RoundTripCallback
}

// RoundTrip implements the http.RoundTripper interface.
func (d *WebsocketRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	referrer := r.Referer()
	if referrer == "" {
		referrer = "http://localhost/"
	}

	wsconf, err := websocket.NewConfig(r.URL.String(), referrer)
	if err != nil {
		return nil, err
	}
	wsconf.TlsConfig = d.TLSConfig
	wsconf.Header = r.Header
	wsconf.Protocol = []string{"channel.k8s.io"}

	conn, err := websocket.DialConfig(wsconf)
	if err != nil {
		return nil, err
	}
	conn.PayloadType = websocket.BinaryFrame
	defer conn.Close()

	return d.Do(conn)
}

// ExecRoundTripper creates a wrapped WebsocketRoundTripper
func ExecRoundTripper(conf *rest.Config, f RoundTripCallback) (http.RoundTripper, error) {
	tlsConfig, err := rest.TLSConfigFor(conf)
	if err != nil {
		return nil, err
	}

	rt := &WebsocketRoundTripper{
		Do:        f,
		TLSConfig: tlsConfig,
	}

	return rest.HTTPWrappersForConfig(conf, rt)
}

// Exec returns an "exec" Request suitable for ExecRoundTripper.
func Exec(client corev1.CoreV1Interface, pod, namespace string, opts v1.PodExecOptions) (*http.Request, error) {
	cl := client.RESTClient()
	req := cl.Verb("ignored").
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(&opts, scheme.ParameterCodec)

	url := req.URL()

	switch url.Scheme {
	case "http":
		url.Scheme = "ws"
	case "https":
		url.Scheme = "wss"
	default:
		return nil, fmt.Errorf("Unrecognised URL scheme in %v", url)
	}

	// NB: Only some fields are honoured by our RoundTrip implementation
	return &http.Request{
		URL: url,
	}, nil
}
