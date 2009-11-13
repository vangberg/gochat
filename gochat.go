package gochat

import (
        "net";
        /*"fmt";*/
        "bufio";
        "container/vector";
)

type Server struct {
        listener *net.TCPListener;
        incomingMessages chan Message;
        registerForMessages chan chan Message;
        outgoingChannels *vector.Vector;
}

type Message struct {
        sender []byte;
        message []byte;
}

func (s *Server) StartServer() {
        s.incomingMessages = make(chan Message);
        s.registerForMessages = make(chan chan Message);
        s.outgoingChannels = new(vector.Vector);

        go s.startRouter();
        s.listenForConnections();
}

func (s *Server) startRouter() {
        for {
                select {
                case msg := <-s.incomingMessages:
                        for i := 0; i < s.outgoingChannels.Len(); i++ {
                                s.outgoingChannels.At(i).(chan Message) <- msg;
                        }
                case newChannel := <-s.registerForMessages:
                        s.outgoingChannels.Push(newChannel);
                }
        }
}

func (s *Server) listenForConnections() {
        ip := net.ParseIP("127.0.0.1");
        addr := &net.TCPAddr{ip, 9999};
        s.listener, _ = net.ListenTCP("tcp", addr);

        for { s.acceptClient(); }
}

func (s *Server) acceptClient() {
        conn, _ := s.listener.AcceptTCP();
        client := newClient(conn, s.incomingMessages);

        client.requestNick();
        s.registerForMessages <- client.outgoingMessages;
        go client.sendReceiveMessages();
}

type Client struct {
        conn *net.TCPConn;
        incomingMessages chan Message;
        outgoingMessages chan Message;
        nickname []byte;
}

func newClient(conn *net.TCPConn, incoming chan Message) (c *Client) {
        return &Client{conn, incoming, make(chan Message)};
}

func (c *Client) requestNick() {
        r := bufio.NewReader(c.conn);

        c.conn.Write("Please enter your nickname: ");

        c.nickname, _ = r.ReadBytes('\n');
}

func (c *Client) sendReceiveMessages() {
        go c.receiveMessages();
        c.sendMessages();
}

func (c *Client) receiveMessages() {
        r := bufio.NewReader(c.conn);
        for {
                bytes, _ := r.ReadBytes('\n');
                msg := Message{c.nickname, bytes};
                c.incomingMessages <-msg;
        }
}

func (c *Client) sendMessages() {
        for {
                msg := <-c.outgoingMessages;
                if msg.nickname != c.nickname {
                        c.conn.Write(msg.message);
                }
        }
}
