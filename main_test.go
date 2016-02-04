package main

import (
	"net"
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(S{})

func (S) TestHandle(c *check.C) {
	apps["myapp.cloud.tsuru.com"] = "myapp"
	defer func() { apps["myapp.cloud.tsuru.com"] = "" }()
	input := []byte("galeb.myapp_cloud_tsuru_com.10_236_99_181_32772.requestTime:44|ms")
	expected := &document{
		Client: "tsuru",
		Metric: "response_time",
		Count:  1,
		App:    "myapp",
		Value:  44,
	}
	doc, err := handle(input)
	c.Assert(err, check.IsNil)
	c.Assert(doc, check.DeepEquals, expected)
}

func runServer(c *check.C) {
	addr, err := net.ResolveUDPAddr("udp", endpoint)
	c.Assert(err, check.IsNil)
	conn, err := net.ListenUDP("udp", addr)
	c.Assert(err, check.IsNil)
	defer conn.Close()
}

func (S) TestSendDocument(c *check.C) {
	apps["myapp.cloud.tsuru.com"] = "myapp"
	defer func() { apps["myapp.cloud.tsuru.com"] = "" }()
	endpoint = ":1984"
	runServer(c)
	doc := &document{
		Client: "tsuru",
		Metric: "response_time",
		Count:  1,
		App:    "myapp",
		Value:  44,
	}
	err := sendDocument(doc)
	c.Assert(err, check.IsNil)
}

func (S) TestAppFromAddr(c *check.C) {
	apps["myapp.com"] = "myapp"
	defer func() { apps["myapp.com"] = "" }()
	c.Assert(appFromAddr("myapp.com"), check.Equals, "myapp")
}

func (S) TestParseAddr(c *check.C) {
	c.Assert(parseAddr("myapp_com"), check.Equals, "myapp.com")
}
