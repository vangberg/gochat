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
	registerForMessages	chan *Client;
	clients			*vector.Vector;
}

type Message struct {
	sender	string;
	message	string;
}

func (s *Server) StartServer() {
	s.incomingMessages = make(chan Message);
	s.registerForMessages = make(chan *Client);
	s.clients = new(vector.Vector);

	go s.listenForConnections();
	s.startRouter();
}

func (s *Server) startRouter() {
	for {
		select {
		case msg := <-s.incomingMessages:
			s.fanOutMessage(msg)
		case newClient := <-s.registerForMessages:
			msg := Message{"System", "New user " +
				newClient.nickname + " has joined.\n"};
			s.fanOutMessage(msg);

			s.clients.Push(newClient);

			newClient.outgoingMessages <- s.clientList();
		}
	}
}

func (s *Server) clientList() (m Message) {
	clientNicknames := make([]string, s.clients.Len());
	for i := 0; i < s.clients.Len(); i++ {
		client := s.clients.At(i).(*Client);
		clientNicknames[i] = client.nickname;
	}

	str := strings.Join(clientNicknames, "\n   ") + "\n";

	return Message{"Online users", str};
}

func (s *Server) fanOutMessage(msg Message) {
	for i := 0; i < s.clients.Len(); i++ {
		client := s.clients.At(i).(*Client);
		ch := client.outgoingMessages;

		if closed(ch) {
			s.clients.Delete(i);
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
	s.registerForMessages <- client;
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
                        msg := Message{"System", "User " + c.nickname + 
                                " has left.\n"};
                        c.incomingMessages <- msg;
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
