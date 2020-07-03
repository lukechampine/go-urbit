package airlock

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) {
	c, err := NewClient("http://localhost:80", "lidlut-tabwed-pillex-ridrup")
	if err != nil {
		t.Skipf("Spin up a fakezod to test (got error: %v)", err)
	}
	s, err := c.Subscribe("zod", "chat-store", "/mailbox/~/~zod/mc")
	if err != nil {
		t.Fatal(err)
	}
	var events []json.RawMessage
	done := make(chan struct{})
	go func() {
		for e := range s.Events {
			events = append(events, e)
		}
		close(done)
	}()

	now := time.Now().Unix() * 1000
	msg := json.RawMessage(fmt.Sprintf(`
{
	"message": {
		"path": "/~/~zod/mc",
		"envelope": {
			"uid": "0v20l.5k520.74net.u5qnm.vlbn4.9on3i.80m7n.6c2fq.s7lsj.l5lcr.8d7q7.klh1f.c6a33.8o75q.pqsh2.kmuqn.5694m.9dg1q.ulkv8.gk8ak.sjobd",
			"number": 1,
			"author": "~zod",
			"when": %v,
			"letter": {"text": "hello there!"}
		}
	}
}`, now))
	if err := c.Poke("zod", "chat-hook", "json", msg); err != nil {
		t.Fatal(err)
	}

	if err := s.Unsubscribe(); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(); err != nil {
		t.Fatal(err)
	}

	<-done
	// should have received an event with our message in it
	var found bool
	for _, e := range events {
		found = found || strings.Contains(string(e), "hello there!")
	}
	if !found {
		t.Fatal("message not found in subscription events")
	}
}
