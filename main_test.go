package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	dep "github.com/hashicorp/consul-template/dependency"
	"github.com/hashicorp/consul-template/test"
	"github.com/hashicorp/consul/sdk/testutil"
)

var (
	testConsul  *testutil.TestServer
	testClients *dep.ClientSet
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	tb := &test.TestingTB{}
	consul, err := testutil.NewTestServerConfigT(tb,
		func(c *testutil.TestServerConfig) {
			c.LogLevel = "warn"
			c.Stdout = ioutil.Discard
			c.Stderr = ioutil.Discard
		})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to start consul server: %v", err))
	}
	testConsul = consul

	clients := dep.NewClientSet()
	if err := clients.CreateConsulClient(&dep.CreateConsulClientInput{
		Address: testConsul.HTTPAddr,
	}); err != nil {
		testConsul.Stop()
		log.Fatal(err)
	}
	testClients = clients

	exitCh := make(chan int, 1)
	func() {
		defer func() {
			// Attempt to recover from a panic and stop the server. If we don't stop
			// it, the panic will cause the server to remain running in the
			// background. Here we catch the panic and the re-raise it.
			if r := recover(); r != nil {
				testConsul.Stop()
				panic(r)
			}
		}()

		exitCh <- m.Run()
	}()

	exit := <-exitCh

	tb.DoCleanup()
	testConsul.Stop()
	os.Exit(exit)
}
