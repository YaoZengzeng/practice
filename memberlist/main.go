package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

var (
	mtx        sync.RWMutex
	members    = flag.String("members", "", "comma seperated list os members")
	port       = flag.Int("port", 4001, "http port")
	bindPort   = flag.Int("bind-port", 8001, "gossip cluster bind port")
	items      = map[string]string{}
	broadcasts *memberlist.TransmitLimitedQueue
)

func addHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")

	mtx.Lock()
	items[key] = val
	mtx.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: "add",
			Data: map[string]string{
				key: val,
			},
		},
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	mtx.Lock()
	delete(items, key)
	mtx.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: "del",
			Data: map[string]string{
				key: "",
			},
		},
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	mtx.RLock()
	val := items[key]
	mtx.RUnlock()

	w.Write([]byte(val))
}

func main() {
	flag.Parse()

	if err := initMemberList(); err != nil {
		fmt.Printf("Init member list failed: %v\n", err)
	}

	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)

	fmt.Printf("Listening on: %d\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Printf("http listen and serve failed: %v\n", err)
	}
}

type update struct {
	Action string
	Data   map[string]string
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

type delegate struct{}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd':
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			fmt.Printf("Unmarshal data to update struct failed: %v\n", err)
			return
		}

		mtx.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case "add":
					items[k] = v
				case "del":
					delete(items, k)
				}
			}
		}
		mtx.Unlock()
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("Marshal in LocalState failed: %v\n", err)
		return nil
	}

	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}

	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		fmt.Printf("Unmarshal in MergeRemoteState failed: %v\n", err)
		return
	}

	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()
}

type eventDelegate struct{}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Printf("A node has joined: %s\n", node.String())
}

func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Printf("A node has left: %s\n", node.String())
}

func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	fmt.Printf("A node was updated: %s\n", node.String())
}

func initMemberList() error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}
	c.Delegate = &delegate{}
	c.BindPort = *bindPort

	rand.Seed(time.Now().UnixNano())
	c.Name = hostname + "-" + strconv.Itoa(rand.Intn(1000))

	m, err := memberlist.Create(c)
	if err != nil {
		return err
	}

	if len(*members) > 0 {
		parts := strings.Split(*members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return err
		}
	}

	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := m.LocalNode()
	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)
	return nil
}
