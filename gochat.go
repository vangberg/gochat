package gochat

import (
	"os";
	"net";
	"fmt";
	"bufio";
	"strings";
	"container/vector";
)

type Server struct {
	listener		*net.TCPListener;
	incomingMessages	chan Message;
	registerForMessages	chan chan Message;
	outgoingChannels	*vector.Vector;
}

type Message struct {
	sender	string;
	message	string;
}

func (s *Server) StartServer() {
	s.incomingMessages = make(chan Message);
	s.registerForMessages = make(chan chan Message);
	s.outgoingChannels = new(vector.Vector);

	go s.listenForConnections();
	s.startRouter();
}

func (s *Server) startRouter() {
	for {
		select {
		case msg := <-s.incomingMessages:
			s.fanOutMessage(msg)
		case newChannel := <-s.registerForMessages:
			s.outgoingChannels.Push(newChannel)
		}
	}
}

func (s *Server) fanOutMessage(msg Message) {
	for i := 0; i < s.outgoingChannels.Len(); i++ {
		ch := s.outgoingChannels.At(i).(chan Message);

		if closed(ch) {
			s.outgoingChannels.Delete(i);
			i--;
		} else {
			ch <- msg
		}
	}
}

func (s *Server) listenForConnections() {
	ip := net.ParseIP("127.0.0.1");
	addr := &net.TCPAddr{ip, 9999};
	s.listener, _ = net.ListenTCP("tcp", addr);

	for {
		s.acceptClient()
	}
}

func (s *Server) acceptClient() {
	conn, _ := s.listener.AcceptTCP();
	client := newClient(conn, s.incomingMessages);

	client.requestNick();
	s.registerForMessages <- client.outgoingMessages;
	client.sendReceiveMessages();
}

type Client struct {
	conn			*net.TCPConn;
	nickname		string;
	incomingMessages	chan Message;
	outgoingMessages	chan Message;
	reader			*bufio.Reader;
}

func newClient(conn *net.TCPConn, incoming chan Message) (c *Client) {
	c = new(Client);
	c.conn = conn;
	c.incomingMessages = incoming;
	c.outgoingMessages = make(chan Message);
	c.reader = bufio.NewReader(c.conn);
	return c;
}

func (c *Client) requestNick() {
	c.conn.Write(strings.Bytes("Please enter your nickname: "));

	nickname, _ := c.reader.ReadString('\n');
	c.nickname = nickname[0 : len(nickname)-1];
}

func (c *Client) sendReceiveMessages() {
	go c.receiveMessages();
	go c.sendMessages();
}

func (c *Client) receiveMessages() {
	for {
		bytes, err := c.reader.ReadString('\n');
		if err == nil {
			msg := Message{c.nickname, bytes};
			c.incomingMessages <- msg;
		}
		if err == os.EOF {
			close(c.outgoingMessages);
			return;
		}
	}
}

func (c *Client) sendMessages() {
	for {
		msg := <-c.outgoingMessages;
		if closed(c.outgoingMessages) {
			return
		}
		if msg.sender != c.nickname {
			c.sendMessage(msg)
		}
	}
}

func (c *Client) sendMessage(msg Message) {
	str := fmt.Sprintf("%s\n   %s", msg.sender, msg.message);
	c.conn.Write(strings.Bytes(str));
}
