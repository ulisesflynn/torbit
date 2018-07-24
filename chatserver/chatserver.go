package chatserver

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	u "github.com/ulisesflynn/torbit/user"
)

// ChatServer contains the configuration needed for the server to run.
type ChatServer struct {
	users   map[string]*u.User
	chatLog io.Writer
	mu      sync.Mutex
}

// NewChatServer creates the chat server.
func NewChatServer(chatLog io.Writer) *ChatServer {
	return &ChatServer{
		users:   make(map[string]*u.User),
		chatLog: chatLog,
		mu:      sync.Mutex{},
	}
}

// Join adds a user to the chat server.
func (cs *ChatServer) Join(user *u.User) error {
	userName := user.Name
	userJoined := fmt.Sprintf("user: %s joined at: %s\n", userName, getTimeStamp())
	_, err := cs.chatLog.Write([]byte(userJoined))
	if err != nil {
		return fmt.Errorf("unable to write msg to chat log, error: %s", err)
	}
	err = cs.writeLoggedInUsers(user)
	if err != nil {
		return err
	}
	cs.mu.Lock()
	cs.users[userName] = user
	cs.mu.Unlock()
	return cs.write(userName, "\n"+userJoined, false)
}

// SendMsg sends a message to all clients.
func (cs *ChatServer) SendMsg(userName, msg string) error {
	sendingMsg := fmt.Sprintf("user: %s sent message: %s\n", userName, msg)
	_, err := cs.chatLog.Write([]byte(sendingMsg))
	if err != nil {
		return fmt.Errorf("unable to write msg to chat log, error: %s", err)
	}
	return cs.write(userName, msg+"\n", true)
}

// Exit removes a user from the chat server.
func (cs *ChatServer) Exit(userName string) error {
	userExited := fmt.Sprintf("user: %s has left at: %s\n", userName, getTimeStamp())
	_, err := cs.chatLog.Write([]byte(userExited))
	if err != nil {
		return fmt.Errorf("unable to write msg to chat log, error: %s", err)
	}
	cs.mu.Lock()
	delete(cs.users, userName)
	cs.mu.Unlock()
	// send exit msg to all clients
	return cs.write(userName, "\n"+userExited, false)
}

// GetUsers returns the current user on the server.
func (cs *ChatServer) GetUsers() map[string]*u.User {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.users
}

// write sends a message to all the user's on the server
func (cs *ChatServer) write(userName, msg string, displayPrompt bool) error {
	var (
		myUser   *u.User
		writeErr error
	)
	cs.mu.Lock()
	defer func() {
		cs.mu.Unlock()
	}()
	for _, usr := range cs.users {
		if usr.Name == userName {
			myUser = usr
			continue
		}
		outMsg := []byte(msg)
		// this displays a prompt with a message
		if displayPrompt {
			outMsg = []byte("\n" + userName + " >> " + msg)
		}
		_, err := usr.Conn.Write(outMsg)
		if err != nil {
			writeErr = errors.Wrap(writeErr, fmt.Sprintf("unable to write msg: %s to usr: %s, error: %s", msg, userName, err))
			continue
		}
		// rewrite all the user's prompts
		_, err = usr.Conn.Write([]byte(usr.Name + " >> "))
		if err != nil {
			writeErr = errors.Wrap(writeErr, fmt.Sprintf("unable to set user prompt for user: %s, error: %s", usr.Name, err))
			continue
		}
	}
	if myUser != nil {
		_, err := myUser.Conn.Write([]byte(myUser.Name + " >> "))
		if err != nil {
			writeErr = errors.Wrap(writeErr, fmt.Sprintf("unable to set user prompt for user: %s, error: %s", myUser.Name, err))
		}
	}
	return writeErr
}

// writeLoggedInUsers displays the users that are currently logged in.
func (cs *ChatServer) writeLoggedInUsers(user *u.User) error {
	loggedInUsers := []byte{}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	numUsers := len(cs.users)
	i := 0
	for _, u := range cs.users {
		i++
		uName := u.Name
		if i < numUsers {
			uName = uName + ", "
		} else {
			uName = uName + "\n"
		}
		loggedInUsers = append(loggedInUsers, []byte(uName)...)
	}
	if len(loggedInUsers) < 1 {
		return nil
	}
	_, err := user.Conn.Write([]byte(fmt.Sprintf("Current logged in users: %s", loggedInUsers)))
	if err != nil {
		return fmt.Errorf("unable to write logged in users, error: %s", err)
	}
	return nil
}

// getTimeStamp returns the current time in UTC.
func getTimeStamp() string {
	const timeFormat = "Mon Jan _2 15:04:05 2006"
	return time.Now().UTC().Format(timeFormat)
}
