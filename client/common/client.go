package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"
	"encoding/binary"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) StartClientBetSending() {
	// This function sends a bet to the server and waits for a response
	c.createClientSocket()
	c.SendBet()
	c.ReceiveBetResponse()
}

func (c *Client) Shutdown() {
	if c.conn != nil {
		c.conn.Close()
		log.Info("action: close_client_socket | result: success")
	}
}

func (c *Client) SafeRead(length int) ([]byte, error) {
	data := make([]byte, length)
	totalRead := 0

	for totalRead < length {
		n, err := c.conn.Read(data[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}

	return data, nil
}

func (c *Client) SafeWrite(data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := c.conn.Write(data[totalSent:])
		if err != nil {
			return err // This means the connection was closed and a short-write occurred
		}
		totalSent += n
	}

	return nil
}

func GenerateMessage() []byte {
	// This function generates a message to be sent to the server
	// 4 bytes for the message length
	// 1 byte for the message type
	// 1 byte for the payload

	clientId := os.Getenv("CLI_ID")
	name := os.Getenv("NAME")
	lastName := os.Getenv("LAST_NAME")
	document := os.Getenv("DOCUMENT")
	birthdate := os.Getenv("BIRTHDATE")
	number := os.Getenv("NUMBER")

	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s", clientId, name, lastName, document, birthdate, number)
	payloadLength := len(payload)
	messageLength := payloadLength + 5 // 4 bytes for length + 1 byte for type

	message := make([]byte, messageLength)
	binary.BigEndian.PutUint32(message[:4], uint32(messageLength))

	// Set message type (0 for bets)
	message[4] = 0
	copy(message[5:], payload)
	return message
}

func (c *Client) SendBet() {
	// This function sends a bet to the server
	// The bet is formed by the GenerateMessage function
	// The message is sent to the server
	message := GenerateMessage()
	c.SafeWrite(message)
}

func (c *Client) ReceiveBetResponse() {
	// This function receives a response for the bet sent to the server
	// The response can be ACK or Error
	// The Response has the format:
	// {type}|{document}|{betAmount}
	// Where type can be 0 for ACK and 1 for Error

	response, err := c.SafeRead(3)
	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	responseType := response[0]
	document := response[1]
	betAmount := response[2]

	if responseType == 0 {
		log.Infof("action: receive_response | result: success | client_id: %v | response: ACK | document: %v | bet_amount: %v",
			c.config.ID,
			document,
			betAmount,
		)
	} else {
		log.Infof("action: receive_response | result: success | client_id: %v | response: Error | document: %v | bet_amount: %v",
			c.config.ID,
			document,
			betAmount,
		)
	}
}