package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"
	"encoding/binary"
	"strings"

	"github.com/op/go-logging"
)

const Delimiter = "|"
const MessageTypeSuccess = 0
const MessageTypeBet = 0
const MessageTypePos = 4
const HeaderLength = 5
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
	messageLength := payloadLength + HeaderLength // 4 bytes for length + 1 byte for type

	message := make([]byte, messageLength)
	binary.BigEndian.PutUint32(message[:MessageTypePos], uint32(messageLength))

	// Set message type (0 for bets)
	message[MessageTypePos] = MessageTypeBet
	copy(message[HeaderLength:], payload)
	return message
}

func (c *Client) SendBet() {
	// This function sends a bet to the server
	// The bet is formed by the GenerateMessage function
	// The message is sent to the server
	message := GenerateMessage()
	c.SafeWrite(message)
}

func (c *Client) ProcessResponse(response []byte) (int, string, string) {
	// This function processes the response received from the server
	// The response has the format:
	// The rest has the format "responseType|document|betAmount"
	// responseType is 0 if the bet was sent correctly or 1 if there was an error

	// decode the response from bytes to string

	decoded_response := string(response)
	parts := strings.Split(decoded_response, Delimiter)
	responseType := int(parts[0][0] - '0') // We have to subtract a string 0 because a string 0 maps to int 48.
	document := parts[1]
	betAmount := parts[2]

	return responseType, document, betAmount
}

func (c *Client) ReceiveBetResponse() {
	// This function receives a response for the bet sent to the server
	// And validates if the bet was sent correctly.
	// It logs the result of the operation

	response_length, err := c.SafeRead(2)
	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	response, err := c.SafeRead(int(binary.BigEndian.Uint16(response_length)))
	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	responseType, document, betAmount := c.ProcessResponse(response)

	if responseType == MessageTypeSuccess {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
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

