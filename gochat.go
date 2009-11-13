package gochat

import (
        "net";
        "fmt";
        "bufio";
        "strings";
        "container/vector";
)

type Server struct {
        listener *net.TCPListener;
        incomingMessages chan Message;
        registerForMessages chan chan Message;
        outgoingChannels *vector.Vector;
}

type Message struct {
        sender string;
        message string;
}

func (s *Server) StartServer() {
        fmt.Println(" * Starting server…");

        s.incomingMessages = make(chan Message);
        s.registerForMessages = make(chan chan Message);
        s.outgoingChannels = new(vector.Vector);

        go s.startRouter();
        s.listenForConnections();
}

func (s *Server) startRouter() {
        fmt.Println(" * Starting message router…");

        for {
                select {
                case msg := <-s.incomingMessages:
                        fmt.Println("Received message: ", msg);
                        for i := 0; i < s.outgoingChannels.Len(); i++ {
                                s.outgoingChannels.At(i).(chan Message) <- msg;
                        }
                case newChannel := <-s.registerForMessages:
                        fmt.Println("Channel registered for messages…");
                        s.outgoingChannels.Push(newChannel);
                }
        }
}

func (s *Server) listenForConnections() {
        fmt.Println(" * Listening for connections…");
        ip := net.ParseIP("127.0.0.1");
        addr := &net.TCPAddr{ip, 9999};
        s.listener, _ = net.ListenTCP("tcp", addr);

        for { s.acceptClient(); }
}

func (s *Server) acceptClient() {
        conn, _ := s.listener.AcceptTCP();
        fmt.Println(" * Accepting client…");
        client := newClient(conn, s.incomingMessages);

        client.requestNick();
        fmt.Println(" * Registering ", client.nickname, " for messages…");
        s.registerForMessages <- client.outgoingMessages;
        go client.sendReceiveMessages();
}

type Client struct {
        conn *net.TCPConn;
        nickname string;
        incomingMessages chan Message;
        outgoingMessages chan Message;
        reader *bufio.Reader;
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
        fmt.Println(" * Requesting nick name from new client…");
        c.conn.Write(strings.Bytes("Please enter your nickname: "));

        nickname, _ := c.reader.ReadString('\n');
        c.nickname = nickname[0:len(nickname)-1];
}

func (c *Client) sendReceiveMessages() {
        fmt.Println(" * Sending and receiving messages for ", c.nickname);
        go c.receiveMessages();
        c.sendMessages();
}

func (c *Client) receiveMessages() {
        for {
                bytes, _ := c.reader.ReadString('\n');
                msg := Message{c.nickname, bytes};
                c.incomingMessages <-msg;
        }
}

func (c *Client) sendMessages() {
        for {
                msg := <-c.outgoingMessages;
                if msg.sender != c.nickname {
                        c.sendMessage(msg);
                }
        }
}

func (c *Client) sendMessage(msg Message) {
        str := fmt.Sprintf("%s\n   %s", msg.sender, msg.message);
        c.conn.Write(strings.Bytes(str));
}
