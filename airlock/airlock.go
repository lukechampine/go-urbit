package airlock

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"lukechampine.com/frand"
)

// A Client facilitates an airlock connection to an Urbit.
type Client struct {
	addr string
	http http.Client
	cond *sync.Cond // for waking goroutines waiting on SSE acks
	once sync.Once  // for spawning sseLoop

	mu          sync.Mutex
	nextID      int
	lastSeenID  int
	lastAckedID int
	subs        map[int]chan json.RawMessage
	acks        map[int]error
	sseErr      error
}

func (c *Client) nextEventID() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	return c.nextID
}

func (c *Client) sseLoop() error {
	req, _ := http.NewRequest("GET", c.addr, nil)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't initiate SSE connection: %w", err)
	}
	defer resp.Body.Close()
	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		line := s.Bytes()
		switch {
		case bytes.HasPrefix(line, []byte("id: ")):
			id, err := strconv.Atoi(string(line[4:]))
			if err != nil {
				return fmt.Errorf("couldn't parse SSE ID: %w", err)
			}
			c.mu.Lock()
			c.lastSeenID = id
			c.mu.Unlock()
		case bytes.HasPrefix(line, []byte("data: ")):
			var data struct {
				ID       int
				Response string
				OK       string
				Err      string
				JSON     json.RawMessage
			}
			if err := json.Unmarshal(line[6:], &data); err != nil {
				return fmt.Errorf("couldn't parse SSE data: %w", err)
			}
			switch data.Response {
			case "subscribe", "poke":
				c.mu.Lock()
				if data.Err == "" {
					c.acks[data.ID] = nil
				} else {
					c.acks[data.ID] = errors.New(data.Err)
				}
				c.mu.Unlock()
				c.cond.Broadcast()
			case "diff":
				c.mu.Lock()
				if s, ok := c.subs[data.ID]; ok {
					s <- data.JSON
				}
				c.mu.Unlock()
			case "quit":
				c.mu.Lock()
				if s, ok := c.subs[data.ID]; ok {
					close(s)
					delete(c.subs, data.ID)
				}
				c.mu.Unlock()
			}
		}
	}
	return s.Err()
}

func (c *Client) waitForAck(id int) error {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	for {
		if c.sseErr != nil {
			return c.sseErr
		}
		if err, ok := c.acks[id]; ok {
			delete(c.acks, id)
			return err
		}
		c.cond.Wait()
	}
}

func (c *Client) sendJSONToChannel(v ...interface{}) error {
	// include ack if necessary
	c.mu.Lock()
	lastAcked, lastSeen := c.lastAckedID, c.lastSeenID
	c.mu.Unlock()
	if lastAcked != lastSeen {
		// the ack MUST come before other messages; if the ack comes after a
		// delete, eyre will get mad at us
		v = append([]interface{}{struct {
			Action  string `json:"action"`
			EventID int    `json:"event-id"`
		}{"ack", lastSeen}}, v...)
	}

	js, _ := json.Marshal(v)
	req, _ := http.NewRequest("PUT", c.addr, bytes.NewReader(js))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return errors.New("invalid request")
		case http.StatusForbidden:
			return errors.New("unauthenticated request (expired cookie?)")
		default:
			return fmt.Errorf("HTTP status code %v (%v)", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
	}

	c.mu.Lock()
	c.lastAckedID = lastSeen
	c.mu.Unlock()

	// initiate SSE connection (if not already connected)
	c.once.Do(func() {
		go func() {
			if err := c.sseLoop(); err != nil {
				c.mu.Lock()
				c.sseErr = err
				c.mu.Unlock()
				c.cond.Broadcast()
			}
		}()
	})
	return nil
}

// Poke sends a poke and waits for it to be acknowledged.
func (c *Client) Poke(ship, app, mark string, v interface{}) error {
	id := c.nextEventID()
	err := c.sendJSONToChannel(struct {
		ID     int         `json:"id"`
		Action string      `json:"action"`
		Ship   string      `json:"ship"`
		App    string      `json:"app"`
		Mark   string      `json:"mark"`
		JSON   interface{} `json:"json"`
	}{id, "poke", ship, app, mark, v})
	if err != nil {
		return err
	}
	return c.waitForAck(id)
}

// Subscribe sets up a subscription on the specified path.
func (c *Client) Subscribe(ship, app, path string) (*Subscription, error) {
	id := c.nextEventID()
	eventCh := make(chan json.RawMessage, 1)
	c.mu.Lock()
	c.subs[id] = eventCh
	c.mu.Unlock()
	err := c.sendJSONToChannel(struct {
		ID     int    `json:"id"`
		Action string `json:"action"`
		Ship   string `json:"ship"`
		App    string `json:"app"`
		Path   string `json:"path"`
	}{id, "subscribe", ship, app, path})
	if err != nil {
		return nil, err
	}
	if err := c.waitForAck(id); err != nil {
		return nil, err
	}
	return &Subscription{
		c:      c,
		id:     id,
		Events: eventCh,
	}, nil
}

// Delete deletes the airlock channel.
func (c *Client) Delete() error {
	return c.sendJSONToChannel(struct {
		ID     int    `json:"id"`
		Action string `json:"action"`
	}{c.nextEventID(), "delete"})
}

// NewClient connects to the Urbit listening on the specified address.
func NewClient(addr, code string) (*Client, error) {
	c := &Client{
		addr: fmt.Sprintf("%v/~/channel/go-airlock-%v", addr, hex.EncodeToString(frand.Bytes(6))),
		acks: make(map[int]error),
		subs: make(map[int]chan json.RawMessage),
	}
	c.cond = sync.NewCond(&c.mu)

	resp, err := c.http.Post(fmt.Sprintf("%v/~/login", addr), "application/x-www-form-urlencoded", strings.NewReader("password="+code))
	if err != nil {
		return nil, err
	}
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ship returned HTTP status code %v (%v)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	c.http.Jar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(addr)
	c.http.Jar.SetCookies(u, resp.Cookies())
	return c, nil
}

// A Subscription is an active subscription.
type Subscription struct {
	c      *Client
	id     int
	Events <-chan json.RawMessage
}

// Unsubscribe unsubscribes from the subscription.
func (s *Subscription) Unsubscribe() error {
	// NOTE: we do not wait for acknowledgement here
	err := s.c.sendJSONToChannel(struct {
		ID           int    `json:"id"`
		Action       string `json:"action"`
		Subscription int    `json:"subscription"`
	}{s.c.nextEventID(), "unsubscribe", s.id})
	if err != nil {
		return err
	}
	s.c.mu.Lock()
	if ch, ok := s.c.subs[s.id]; ok {
		close(ch)
		delete(s.c.subs, s.id)
	}
	s.c.mu.Unlock()
	return nil
}
