// Copyright 2014 docker-cluster authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

/*
import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

)

func TestNewCluster(t *testing.T) {
	var tests = []struct {
		input []Node
		fail  bool
	}{
		{
			[]Node{{Address: "http://localhost:8083"}},
			false,
		},
		{
			[]Node{{Address: ""}, {Address: "http://localhost:8083"}},
			true,
		},
		{
			[]Node{{Address: "http://localhost:8083"}},
			false,
		},
	}
	for _, tt := range tests {
		_, err := New(nil, &MapStorage{}, tt.input...)
		if tt.fail && err == nil || !tt.fail && err != nil {
			t.Errorf("cluster.New() for input %#v. Expect failure: %v. Got: %v.", tt.input, tt.fail, err)
		}
	}
}

func TestNewFailure(t *testing.T) {
	_, err := New(&roundRobin{}, nil)
	if err != errStorageMandatory {
		t.Fatalf("expected errStorageMandatory error, got: %#v", err)
	}
}

func TestRegister(t *testing.T) {
	scheduler := &roundRobin{}
	cluster, err := New(scheduler, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{
		Address:  "http://localhost1:4243",
		Metadata: map[string]string{"x": "y", "a": "b"},
	}
	err = cluster.Register(node)
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.CreateContainerOptions{}
	node, err = scheduler.Schedule(cluster, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if node.Address != "http://localhost1:4243" {
		t.Errorf("Register failed. Got wrong Address. Want %q. Got %q.", "http://localhost1:4243", node.Address)
	}
	err = cluster.Register(Node{Address: "http://localhost2:4243"})
	if err != nil {
		t.Fatal(err)
	}
	node, err = scheduler.Schedule(cluster, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if node.Address != "http://localhost2:4243" {
		t.Errorf("Register failed. Got wrong ID. Want %q. Got %q.", "http://localhost2:4243", node.Address)
	}
	node, err = scheduler.Schedule(cluster, opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if node.Address != "http://localhost1:4243" {
		t.Errorf("Register failed. Got wrong ID. Want %q. Got %q.", "http://localhost1:4243", node.Address)
	}
}

func TestRegisterDoesNotAllowRepeatedAddresses(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://localhost1:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://localhost1:4243"})
	if err != storage.ErrDuplicatedNodeAddress {
		t.Fatalf("Expected error ErrDuplicatedNodeAddress, got: %#v", err)
	}
}

func TestRegisterFailure(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{})
	if err == nil {
		t.Error("Expected non-nil error, got <nil>.")
	}
}

func TestUpdateNode(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{Address: "http://localhost1:4243"}
	err = cluster.Register(node)
	if err != nil {
		t.Fatal(err)
	}
	node.Metadata = map[string]string{"k1": "v1", "k2": "v2"}
	node, err = cluster.UpdateNode(node)
	if err != nil {
		t.Fatal(err)
	}
	expected := Node{Address: "http://localhost1:4243", Metadata: map[string]string{
		"k1": "v1",
		"k2": "v2",
	}}
	node.Healing = HealingData{}
	if !reflect.DeepEqual(node, expected) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", expected, node)
	}
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	nodes[0].Healing = HealingData{}
	if !reflect.DeepEqual(nodes, []Node{expected}) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", []Node{expected}, nodes)
	}
}

func TestUpdateNodeCreationStatus(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{Address: "http://localhost1:4243", CreationStatus: NodeCreationStatusPending}
	err = cluster.Register(node)
	if err != nil {
		t.Fatal(err)
	}
	node.CreationStatus = NodeCreationStatusError
	_, err = cluster.UpdateNode(node)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := cluster.UnfilteredNodes()
	if err != nil {
		t.Fatal(err)
	}
	if nodes[0].CreationStatus != NodeCreationStatusError {
		t.Errorf("UpdateNode: wrong status. Want NodeCreationStatusError. Got %s", nodes[0].CreationStatus)
	}
	node.CreationStatus = NodeCreationStatusPending
	_, err = cluster.UpdateNode(node)
	if err == nil || err.Error() != `cannot update node status when current status is "error"` {
		t.Errorf("UpdateNode: unexpected error %v", err)
	}
}

func TestUpdateNodeRemoveMetadata(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	node := Node{
		Address:  "http://localhost1:4243",
		Metadata: map[string]string{"k1": "v1", "k2": "v2"},
	}
	err = cluster.Register(node)
	if err != nil {
		t.Fatal(err)
	}
	node.Metadata = map[string]string{"k1": "", "k2": "v9", "k3": "v10"}
	node, err = cluster.UpdateNode(node)
	if err != nil {
		t.Fatal(err)
	}
	expected := Node{Address: "http://localhost1:4243", Metadata: map[string]string{
		"k2": "v9",
		"k3": "v10",
	}}
	node.Healing = HealingData{}
	if !reflect.DeepEqual(node, expected) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", expected, node)
	}
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	nodes[0].Healing = HealingData{}
	if !reflect.DeepEqual(nodes, []Node{expected}) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", []Node{expected}, nodes)
	}
}

func TestUpdateNodeStress(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://localhost1:4243"})
	if err != nil {
		t.Fatal(err)
	}
	var errCount int32
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			node := Node{
				Address:  "http://localhost1:4243",
				Metadata: map[string]string{fmt.Sprintf("k%d", i): fmt.Sprintf("v%d", i)},
			}
			_, err := cluster.UpdateNode(node)
			if err == errHealerInProgress {
				atomic.AddInt32(&errCount, 1)
			} else if err != nil {
				t.Fatal(err)
			}
		}(i)
	}
	wg.Wait()
	if errCount <= 0 {
		t.Error("Expected errCount to me greater than 0")
	}
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes[0].Metadata) == 0 {
		t.Error("Expected to have at least one metadata, got 0")
	}
}

func TestUnregister(t *testing.T) {
	scheduler := &roundRobin{}
	cluster, err := New(scheduler, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://localhost1:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Unregister("http://localhost1:4243")
	if err != nil {
		t.Fatal(err)
	}
	opts := docker.CreateContainerOptions{}
	_, err = scheduler.Schedule(cluster, opts, nil)
	if err == nil || err.Error() != "No nodes available" {
		t.Fatal("Expected no nodes available error")
	}
}

func TestNodesShouldGetClusterNodes(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://localhost:4243"})
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Unregister("http://localhost:4243")
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	expected := []Node{{Address: "http://localhost:4243", Metadata: map[string]string{}}}
	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", expected, nodes)
	}
}

func TestNodesShouldGetClusterNodesWithoutDisabledNodes(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	stopChan := make(chan bool)
	healer := &blockingHealer{stop: stopChan}
	defer close(stopChan)
	cluster.Healer = healer
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Unregister("http://server1:4243")
	defer cluster.Unregister("http://server2:4243")
	err = cluster.Register(Node{Address: "http://server1:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://server2:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.handleNodeError("http://server1:4243", errors.New("some err"), true)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan bool)
	go func() {
		stopChan <- true
		for {
			node, err := cluster.storage().RetrieveNode("http://server1:4243")
			if err != nil {
				t.Fatal(err)
			}
			if !node.isHealing() {
				break
			}
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for healer call being made and unlocked")
	}
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	expected := []Node{
		{Address: "http://server2:4243", Metadata: map[string]string{}},
	}
	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("Expected nodes to be equal %#v, got %#v", expected, nodes)
	}
}

func TesteUnfilteredNodesReturnAllNodes(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Unregister("http://server1:4243")
	defer cluster.Unregister("http://server2:4243")
	err = cluster.Register(Node{Address: "http://server1:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{Address: "http://server2:4243"})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.handleNodeError("http://server1:4243", errors.New("some err"), true)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := cluster.UnfilteredNodes()
	if err != nil {
		t.Fatal(err)
	}
	expected := []Node{
		{Address: "http://server1:4243", Metadata: map[string]string{}},
		{Address: "http://server2:4243", Metadata: map[string]string{}},
	}
	sort.Sort(NodeList(nodes))
	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", expected, nodes)
	}
}

func TestNodesForMetadataShouldGetClusterNodesWithMetadata(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{
		Address:  "http://server1:4243",
		Metadata: map[string]string{"key1": "val1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = cluster.Register(Node{
		Address:  "http://server2:4243",
		Metadata: map[string]string{"key1": "val2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Unregister("http://server1:4243")
	defer cluster.Unregister("http://server2:4243")
	nodes, err := cluster.NodesForMetadata(map[string]string{"key1": "val2"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []Node{{Address: "http://server2:4243", Metadata: map[string]string{"key1": "val2"}}}
	if !reflect.DeepEqual(nodes, expected) {
		t.Errorf("Expected nodes to be equal %+v, got %+v", expected, nodes)
	}
}

func TestNodesShouldReturnEmptyListWhenNoNodeIsFound(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := cluster.Nodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 0 {
		t.Errorf("Expected nodes to be empty, got %+v", nodes)
	}
}

func TestRunOnNodesWhenReceiveingNodeShouldntLoadStorage(t *testing.T) {
	id := "e90302"
	body := fmt.Sprintf(`{"Id":"%s","Path":"date","Args":[]}`, id)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := cluster.runOnNodes(func(n node) (interface{}, error) {
		return n.InspectContainer(id)
	}, &docker.NoSuchContainer{ID: id}, false, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	container := result.(*docker.Container)
	if container.ID != id {
		t.Errorf("InspectContainer(%q): Wrong ID. Want %q. Got %q.", id, id, container.ID)
	}
	if container.Path != "date" {
		t.Errorf("InspectContainer(%q): Wrong Path. Want %q. Got %q.", id, "date", container.Path)
	}
}

func TestRunOnNodesStress(t *testing.T) {
	n := 1000
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(16))
	body := `{"Id":"e90302","Path":"date","Args":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()
	id := "e90302"
	cluster, err := New(nil, &MapStorage{}, Node{Address: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < rand.Intn(10)+n; i++ {
		result, err := cluster.runOnNodes(func(n node) (interface{}, error) {
			return n.InspectContainer(id)
		}, &docker.NoSuchContainer{ID: id}, false)
		if err != nil {
			t.Fatal(err)
		}
		container := result.(*docker.Container)
		if container.ID != id {
			t.Errorf("InspectContainer(%q): Wrong ID. Want %q. Got %q.", id, id, container.ID)
		}
		if container.Path != "date" {
			t.Errorf("InspectContainer(%q): Wrong Path. Want %q. Got %q.", id, "date", container.Path)
		}
	}
}

func TestClusterNodes(t *testing.T) {
	c, err := New(&roundRobin{}, &MapStorage{})
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	nodes := []Node{
		{Address: "http://localhost:8080", Metadata: map[string]string{}},
		{Address: "http://localhost:8081", Metadata: map[string]string{}},
	}
	for _, n := range nodes {
		c.Register(n)
	}
	got, err := c.Nodes()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(got, nodes) {
		t.Errorf("roundRobin.Nodes(): wrong result. Want %#v. Got %#v.", nodes, got)
	}
}

func TestClusterNodesUnregister(t *testing.T) {
	c, err := New(&roundRobin{}, &MapStorage{})
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	nodes := []Node{
		{Address: "http://localhost:8080"},
		{Address: "http://localhost:8081"},
	}
	for _, n := range nodes {
		c.Register(n)
	}
	c.Unregister(nodes[0].Address)
	got, err := c.Nodes()
	if err != nil {
		t.Error(err)
	}
	expected := []Node{{Address: "http://localhost:8081", Metadata: map[string]string{}}}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("roundRobin.Nodes(): wrong result. Want %#v. Got %#v.", nodes, got)
	}
}

type blockingHealer struct {
	calls         int
	disabledUntil string
	failureCount  int
	stop          <-chan bool
}

func (h *blockingHealer) HandleError(n *Node) time.Duration {
	h.calls++
	h.failureCount = n.FailureCount()
	h.disabledUntil = n.Metadata["DisabledUntil"]
	<-h.stop
	return 1 * time.Minute
}

func isDateSameMinute(dt1, dt2 string) bool {
	re := regexp.MustCompile(`(.*T\d{2}:\d{2}).*`)
	dt1Minute := re.ReplaceAllString(dt1, "$1")
	dt2Minute := re.ReplaceAllString(dt2, "$1")
	return dt1Minute == dt2Minute
}

func TestClusterHandleNodeErrorStress(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(100))
	c, err := New(&roundRobin{}, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	stopChan := make(chan bool)
	healer := &blockingHealer{stop: stopChan}
	c.Healer = healer
	err = c.Register(Node{Address: "stress-addr-1"})
	if err != nil {
		t.Fatal(err)
	}
	expectedErr := errors.New("some error")
	for i := 0; i < 200; i++ {
		c.handleNodeError("stress-addr-1", expectedErr, true)
	}
	done := make(chan bool)
	go func() {
		stopChan <- true
		for {
			node, err := c.storage().RetrieveNode("stress-addr-1")
			if err != nil {
				continue
			}
			if !node.isHealing() {
				break
			}
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for node unlock")
	}
	if healer.failureCount != 1 {
		t.Errorf("Expected %d failures count, got: %d", 1, healer.failureCount)
	}
	if healer.calls != 1 {
		t.Errorf("Expected healer to have 1 call, got: %d", healer.calls)
	}
	err = c.handleNodeError("stress-addr-1", expectedErr, true)
	if err != nil {
		t.Fatal(err)
	}
	done = make(chan bool)
	go func() {
		stopChan <- true
		for {
			node, err := c.storage().RetrieveNode("stress-addr-1")
			if err != nil {
				continue
			}
			if !node.isHealing() {
				break
			}
		}
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for node unlock")
	}
	if healer.calls != 2 {
		t.Errorf("Expected healer to have 2 calls, got: %d", healer.calls)
	}
	if healer.failureCount != 2 {
		t.Errorf("Expected %d failures count, got: %d", 2, healer.failureCount)
	}
	disabledStr := healer.disabledUntil
	now := time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	if !isDateSameMinute(disabledStr, now) {
		t.Errorf("Expected DisabledUntil to be like %s, got: %s", now, disabledStr)
	}
	nodes, err := c.storage().RetrieveNodes()
	node := nodes[0]
	if err != nil {
		t.Fatal(err)
	}
	if node.FailureCount() != 2 {
		t.Errorf("Expected FailureCount to be 2, got: %d", node.FailureCount())
	}
	if !isDateSameMinute(node.Metadata["DisabledUntil"], disabledStr) {
		t.Errorf("Expected DisabledUntil to be like %s, got: %s", disabledStr, node.Metadata["DisabledUntil"])
	}
}



func TestWrapError(t *testing.T) {
	err := errors.New("my error")
	node := node{addr: "199.222.111.10"}
	wrapped := wrapError(node, err)
	expected := "error in docker node \"199.222.111.10\": my error"
	if wrapped.Error() != expected {
		t.Fatalf("Expected to receive %s, got: %s", expected, wrapped.Error())
	}
	nodeErr, ok := wrapped.(DockerNodeError)
	if !ok {
		t.Fatalf("Expected wrapped to be DockerNodeError")
	}
	if nodeErr.BaseError() != err {
		t.Fatalf("Expected BaseError to be original error")
	}
}

func TestWrapErrorNil(t *testing.T) {
	node := node{addr: "199.222.111.10"}
	wrapped := wrapError(node, nil)
	if wrapped != nil {
		t.Fatalf("Expected to receive nil, got: %#v", wrapped)
	}
}

func TestClusterGetNodeByAddr(t *testing.T) {
	cluster, err := New(nil, &MapStorage{})
	if err != nil {
		t.Fatal(err)
	}
	node, err := cluster.getNodeByAddr("http://199.222.111.10")
	if err != nil {
		t.Fatal(err)
	}
	if node.HTTPClient != cluster.timeout10Client {
		t.Fatalf("Expected client %#v, got %#v", cluster.timeout10Client, node.HTTPClient)
	}
}
*/
