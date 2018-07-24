package user

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

// User contains the name and connection for a user's chat session.
type User struct {
	Name string
	Conn net.Conn
}

// NewUser returns a new user for use by the chat server.
func NewUser(conn net.Conn, connectedUsers map[string]*User) (user *User, err error) {
	var name string
	_, err = io.WriteString(conn, "Please enter your name: ")
	if err != nil {
		return user, err
	}
	scanner := bufio.NewScanner(conn)
	w := bufio.NewWriter(conn)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return user, err
		}
		name = scanner.Text()
		if name == "" {
			_, err = conn.Write([]byte("Name cannot be blank, please try again: "))
			if err != nil {
				return user, err
			}
			continue
		}
		if _, ok := connectedUsers[name]; !ok {
			break
		}
		fmt.Fprintf(w, fmt.Sprintf("The name: %s is already in use, try again", name))
		continue
	}
	return &User{
		Name: name,
		Conn: conn,
	}, nil
}
