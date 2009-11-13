package gochat

import (
        "net";
        "fmt";
        "bufio";
        "container/vector";
)

type Server struct {
        listener *net.TCPListener;
        router *Router;
}

func (s *Server) Listen() {
        ip := net.ParseIP("127.0.0.1");
        addr := &net.TCPAddr{ip, 9999};
        listener, _ := net.ListenTCP("tcp", addr);
        s.listener = listener;

        incoming := make(chan []byte);
        outgoingAdder := make(chan chan []byte);
        router := NewRouter(incoming, outgoingAdder);
        go router.RouteMessages();

        for {
                c, _ := s.listener.AcceptTCP();
                client := newClient(c, incoming);
                outgoingAdder <- client.outgoing;
                go client.ReadMessages();
                go client.WriteMessages();
        }
}

type Client struct {
        conn *net.TCPConn;
        incoming chan []byte;
        outgoing chan []byte;
}

func newClient(conn *net.TCPConn, incoming chan []byte) (c *Client) {
        client := new(Client);
        client.conn = conn;
        client.incoming = incoming;
        client.outgoing = make(chan []byte);
        return client;
}

func (c *Client) ReadMessages() {
        r := bufio.NewReader(c.conn);
        for {
                bytes, _ := r.ReadBytes('\n');
                c.incoming <-bytes;
        }
}

func (c *Client) WriteMessages() {
        for {
                bytes := <-c.outgoing;
                c.conn.Write(bytes);
        }
}

type Router struct {
        incomingMessages chan []byte;
        outgoingAdder chan chan []byte;
        outgoingChannels *vector.Vector;
}

func NewRouter(incoming chan []byte, outgoingAdder chan chan []byte) (r *Router) {
        r = new(Router);
        r.incomingMessages = incoming;
        r.outgoingAdder = outgoingAdder;
        r.outgoingChannels = new(vector.Vector);
        return r
}

func (r *Router) RouteMessages() {
        for {
                select {
                case msg := <-r.incomingMessages:
                        for i := 0; i < r.outgoingChannels.Len(); i++ {
                                r.outgoingChannels.At(i).(chan []byte) <- msg;
                        }
                case ch := <-r.outgoingAdder:
                        r.outgoingChannels.Push(ch);
                        fmt.Print(r.outgoingChannels.Len());
                }
        }
}
