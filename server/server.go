package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cs "github.com/ulisesflynn/torbit/chatserver"
	u "github.com/ulisesflynn/torbit/user"
)

const (
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
)

// New returns a server instance that initializes
// the chat server.
func New(chatLog io.Writer, chatPort, httpPort, srvAddr string, maxMsgSize int) *Server {
	return &Server{
		chatServer: cs.NewChatServer(chatLog),
		chatPort:   chatPort,
		httpPort:   httpPort,
		srvAddr:    srvAddr,
	}
}

// Server contains the fields needed to run the chat server.
type Server struct {
	chatServer   *cs.ChatServer
	chatPort     string
	httpPort     string
	srvAddr      string
	maxMsgSize   int // in bytes
	chatServerUp bool
}

// ServeHTTP handles http requests to the chat server.
func (s *Server) ServeHTTP() {
	r := mux.NewRouter()
	// enable encoded path to handle %2f in path
	r.UseEncodedPath()
	r.Handle("/ping", http.HandlerFunc(s.Ping)).Methods("GET")
	r.Handle("/health_check", http.HandlerFunc(s.HealthCheck)).Methods("GET")
	r.Handle("/send_msg/{username}", http.HandlerFunc(s.SendMsg)).Methods("POST")
	// handle timeouts
	muxWithMiddlewares := http.TimeoutHandler(r, writeTimeout, "client server timeout")
	srv := http.Server{
		Addr:        s.srvAddr + ":" + s.httpPort,
		Handler:     muxWithMiddlewares,
		ReadTimeout: readTimeout,
	}
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

// ServeTelnet listens for incoming connections for the given port.
func (s *Server) ServeTelnet() {
	l, err := net.Listen("tcp", s.srvAddr+":"+s.chatPort)
	if err != nil {
		panic(fmt.Sprintf("unable to listen to chat port, error: %s", err))
	}
	defer l.Close()
	s.chatServerUp = true
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("unable to accept a connection, error: %s\n", err)
			s.chatServerUp = false
			conn.Close()
			return
		}
		go s.manageConn(conn)
	}

}

// manageConn handles all the incoming server connections.
func (s *Server) manageConn(conn net.Conn) {
	defer conn.Close()
	nu, err := u.NewUser(conn, s.chatServer.GetUsers())
	if err != nil {
		log.Printf("unable to create a new user, error: %s\n", err)
		return
	}
	err = s.chatServer.Join(nu)
	if err != nil {
		log.Printf("unable to join chat server, error: %s\n", err)
		return
	}
	defer func() {
		err := s.chatServer.Exit(nu.Name)
		if err != nil {
			log.Printf("unable to exit chat server, error: %s\n", err)
		}
	}()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		err := s.chatServer.SendMsg(nu.Name, msg)
		if err != nil {
			log.Printf("unable to send msg: %s to user: %s", msg, nu.Name)
			return
		}
	}
}

// Run starts the telnet and http servers.
func (s *Server) Run() {
	go s.ServeHTTP()
	go s.ServeTelnet()
}

// Ping endpoint returns 200 OK.
func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong")
}

// HealthCheck returns 200 OK if the underlying chat server is still up.
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if !s.chatServerUp {
		fmt.Fprintln(w, "chat server is down")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

// SendMsg receives the username through a path variable and the msg to send from the request body.
func (s *Server) SendMsg(w http.ResponseWriter, r *http.Request) {
	msg := make([]byte, 0, s.maxMsgSize)
	vars := mux.Vars(r)
	userName := vars["username"]
	if userName == "" {
		fmt.Fprintln(w, "username cannot be blank")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buf := bytes.NewBuffer(msg)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		fmt.Fprintf(w, "unable to read request body, error: %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	usrMsg := buf.String()
	err = s.chatServer.SendMsg(userName, usrMsg)
	if err != nil {
		fmt.Fprintf(w, "unable to send msg: %s from user: %s, error: %s\n", usrMsg, usrMsg, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}
