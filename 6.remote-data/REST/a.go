package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// getReq is the request sent to server to get quote of the day.
type getReq struct {
	// Author is the author you want, if empty it will be a random one
	Author string `json:"author"`
}

// fromReader reads from an io.Reader and unmarshals the content into getReq{},
// This is used to decode from the http.Request.Body into our struct
func (g *getReq) fromReader(r io.Reader) error {
	log.Println("<fromReader>")
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	log.Println(b)
	return json.Unmarshal(b, g)
}

// getResp is response for quote of the day.
type getResp struct {
	// Quote from the server
	Quote string `json:"quote"`
	// Error if a non-http related error.
	Error *Error `json:"error"`
}

// ErrCode is a code so the user can tell what the specific err condition was.
type ErrCode string

// our custom error type for this package
type Error struct {
	Code ErrCode
	Msg  string
}

// Error implements error.Error().
func (e Error) Error() string {
	return fmt.Sprintf("(code %v): %s", e.Code, e.Msg)
}

const (
	UnknownCode   ErrCode = ""
	UnknownAuthor ErrCode = "UnknownAuthor"
)

/*
REST CLIENT
*/

// QOTD represents our client to talk to QOTD server.
type QOTD struct {
	// the URL for the servers address, aka http://someserver.com:80
	u *url.URL
	// this is the *http.Client that will be reused to contact the server
	client *http.Client
}

// New constructs a new QOTD client.
func New(addr string) (*QOTD, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	log.Println("<New>: ", u)
	return &QOTD{
		u:      u,
		client: &http.Client{},
	}, nil
}

// restCall provides a generic POST and JSON REST call function, this can be reused
// with other endpoints
func (q *QOTD) restCall(ctx context.Context, endpoint string, req, resp interface{}) error {
	// if we dont have a deadline we apply a default.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}
	// convert our req into json
	b, err := json.Marshal(req)
	log.Println("<restCall>: ", b)
	if err != nil {
		return err
	}

	// create a new HTTP request using POST  to out endpoint with the body
	// set to our json request.
	hReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(b),
	)
	if err != nil {
		return err
	}
	log.Println("hReq: ", hReq)

	// Make the request
	hResp, err := q.client.Do(hReq)
	if err != nil {
		return err
	}

	// read the response's body
	b, err = io.ReadAll(hResp.Body)
	if err != nil {
		return err
	}

	// unmarshal the json resp into the response
	return json.Unmarshal(b, resp)
}

// Get fetches a quote of the day from the server
func (q *QOTD) Get(ctx context.Context, author string) (string, error) {
	const endpoint = `/qotd/v1/get`
	ref, _ := url.Parse(endpoint)
	resp := getResp{}

	// Makes a call to the server. the endpoint is the joining of our base
	// url (http://127.0.0.1:80) with our constant endpoint abose to form :
	// `http://127.0.0.1:80/qotd/v1/get`
	err := q.restCall(ctx, q.u.ResolveReference(ref).String(), getReq{Author: author}, &resp)
	switch {
	case err != nil: // http error
		return "", err
	case resp.Error != nil: // server error, such as the author not being found
		return "", resp.Error
	}
	return resp.Quote, nil
}

/*
REST SERVER
*/

// server is a REST server for serving quotes of the day
type server struct {
	// serv is the http server we will use.
	serv *http.Server
	// quotes has keys that are names and values that are list of quotes attributed
	quotes map[string][]string
}

// newServer is the constructor for server. The port is the port to run on.
func newServer(port int) (*server, error) {
	s := &server{
		serv: &http.Server{
			Addr: ":" + strconv.Itoa(port), // results in string like ":80"
		},
		quotes: map[string][]string{
			"Mark Twain": {
				"History doesn't repeat itself, but is does ryme",
				"Lies, damned lies and statistic",
				"Gold is a good walk spoiled",
			},
			"Benjamin Franklin": {
				"Tell me and I forget. Teach me and I remember. Involve me and I learn",
				"I didn't fail the test. I just found 180 ways to do it wrong",
			},
			"Eleanor Roosvelt": {
				"The future belongs to those who believe in the beauty of their dreams",
			},
		},
	}
	// A mux handles looking at an incoming URL and determining what function should handle it.
	// This has rules for pattern matching, more reading in: https://pkg.go.dev/net/http#ServerMux
	mux := http.NewServeMux()
	mux.HandleFunc(`/qotd/v1/get`, s.qotdGet)

	// the muxer implements http.Handler and we assign it to our servers URL handling.
	s.serv.Handler = mux

	return s, nil
}

// start starts our server.
func (s *server) start() error {
	return s.serv.ListenAndServe()
}

// qotdGet provides an http.HendleFunc for receiving REST requests for a quote of the day
func (s *server) qotdGet(w http.ResponseWriter, r *http.Request) {
	// Get the Context for the request.
	ctx := r.Context()

	// If no deadline is set, set one.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}

	// read our http.Request's body as JSON into our request object.
	req := getReq{}
	if err := req.fromReader(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var quotes []string

	// no author was requested so we will random
	if req.Author == "" {
		// to get a value from a map, you must know the key.
		// since we are trying to get a randim quote from a random author
		// we will simply do a single loop using range that extracts from the map in random order
		for _, quotes = range s.quotes {
			break
		}
	} else { // auhtor was requested
		// find the authors.
		var ok bool
		quotes, ok = s.quotes[req.Author]
		// no author was found, send a custom error message back.
		if !ok {
			b, err := json.Marshal(
				getResp{
					Error: &Error{
						Code: UnknownAuthor,
						Msg:  fmt.Sprintf("Author %q was not found", req.Author),
					},
				},
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
			return
		}
	}

	// This chooses a random number whose maximum value is the length of our quotes slice.
	// Note that `math/rand` calls vs `crypto/rand` are not cryptographically secure.
	i := rand.Intn(len(quotes))

	// Send our quote back to the client.
	b, err := json.Marshal(getResp{Quote: quotes[i]})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(b)
	return
}

func main() {
	// Sets us some randomization between runs.
	// #rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println("aha")

	// Create a new server listening on port 80. Will listen on all available IP addresses.
	serv, err := newServer(8009)
	if err != nil {
		fmt.Println(1)
		panic(err)
	}
	log.Println(serv)
	// Start our server. This blocks, so we have it do it in its own goroutine.
	go serv.start()
	log.Println("started")

	// Sleep long enought for the server to start.
	time.Sleep(500 * time.Millisecond)

	// Create a client that is pointed at our localhost address on port 80.
	client, err := New("http://127.0.0.1:8009")
	//client, err := New("http://127.0.0.1:8009/qotd/v1/get")
	if err != nil {
		fmt.Println(2)
		panic(err)
	}

	// we are goig to fetch several responses currently and put them in this channel
	results := make(chan string, 2)

	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Get a quote from Mark Twain. He has the best quotes.
	wg.Add(1)
	go func() {
		defer wg.Done()
		quote, err := client.Get(ctx, "Mark Twain")
		if err != nil {
			fmt.Println(3)
			panic(err)
		}
		results <- quote
	}()

	// When we have finished getting quotes, close our results channel.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Read the returned quotes until channel is closed.
	for result := range results {
		fmt.Println(result)
	}
}
