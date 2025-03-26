package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"
	"encoding/binary"
	"strings"
	"strconv"

	"github.com/op/go-logging"
)

const Delimiter = "|"
const MessageTypeSuccess = 0
const MessageTypeBet = 0
const MessageTypeMultipleBet = 1
const MessageTypeReadyForLottery = 2
const MessageTypeAwaitingLottery = 3
const MessageTypePos = 4
const HeaderLength = 5
const BetDelimiter = "\n"
var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchMaxAmount int
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

func (c *Client) StartBettingLoop() {
	// This functions starts a loop that will end when all bets are sent.
	// In each iteration it sends c.config.BatchMaxAmount bets to the server
	// and waits for a response

	file, err := os.Open("./client_bets.csv")
	if err != nil {
		log.Errorf("action: open_file | result: fail | error: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var bets []string

	for scanner.Scan() {
		bets = append(bets, scanner.Text())

		if len(bets) == c.config.BatchMaxAmount {
			c.SendMultipleBets(bets)
			bets = nil // Reset the batch
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("action: read_file | result: fail | error: %v", err)
		return
	}

	if len(bets) > 0 {
		c.SendMultipleBets(bets)
	}

	// Now we have finished sending all the bets
	// We can send a message to the server to let it know that we are ready for the lottery
	c.sendReadyForLottery()

	// We can now wait for the lottery results, we do this 
	// by sending a message to the server and waiting for a response
	for {
		c.createClientSocket()
		c.sendLotteryFinished()

		lotteryFinished, winnersAmount := c.receiveLotteryFinished()
		c.conn.Close()
		if lotteryFinished {
			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", winnersAmount)
			break
		}
		time.Sleep(c.config.LoopPeriod)
	}
	time.Sleep(c.config.LoopPeriod)
}

func (c *Client) SendMultipleBets(bets []string) {
	// This function sends a bet to the server and waits for a response

	c.createClientSocket()
	defer c.conn.Close()
	
	packet := c.CreateBetsPacket(bets)
	err := c.SafeWrite(packet)
	if err != nil {
		log.Errorf("action: send_packet | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	c.ReceiveMultipleBetResponse()
}

func (c *Client) sendReadyForLottery() {
	// This function sends a message to the server to let it know that we are ready for the lottery
	// The message has the format:
	// 4 bytes for the message length
	// 1 byte for the message type
	// Payload: <clientId>

	clientId := os.Getenv("CLI_ID")
	payload := fmt.Sprintf("%s", clientId)
	payloadLength := len(payload)
	messageLength := payloadLength + HeaderLength
	
	message := make([]byte, messageLength)
	binary.BigEndian.PutUint32(message[:MessageTypePos], uint32(messageLength))
	message[MessageTypePos] = MessageTypeReadyForLottery
	copy(message[HeaderLength:], payload)

	c.createClientSocket()
	defer c.conn.Close()
	err := c.SafeWrite(message)
	if err != nil {
		log.Errorf("action: send_ready_for_lottery | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	c.ReceiveReadyForLotteryResponse()
}

func (c *Client) sendLotteryFinished() {
	// This function sends a message to the server querying
	// to see if the lottery has finished
	// The message has the format:
	// 4 bytes for the message length
	// 1 byte for the message type
	// Payload: <clientId>

	clientId := os.Getenv("CLI_ID")
	payload := fmt.Sprintf("%s", clientId)
	payloadLength := len(payload)
	messageLength := payloadLength + HeaderLength

	message := make([]byte, messageLength)
	binary.BigEndian.PutUint32(message[:MessageTypePos], uint32(messageLength))
	message[MessageTypePos] = MessageTypeAwaitingLottery
	copy(message[HeaderLength:], payload)

	err := c.SafeWrite(message)
	if err != nil {
		log.Errorf("action: send_lottery_finished | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
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

func GeneratePayload(clientId, name, lastName, document, birthdate, number string) string {
	// This function generates a payload for a bet
	// It takes the fields as parameters and outputs a formatted string
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", clientId, name, lastName, document, birthdate, number)
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

	payload := GeneratePayload(clientId, name, lastName, document, birthdate, number)
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

func (c *Client) ProcessResponseSingleBet(response []byte) (int, string, string) {
	// This function processes the response received from the server
	// The response has the format:
	// First byte is the message type
	// The rest has the format "responseType|document|betAmount"
	// responseType is 0 if the bet was sent correctly or 1 if there was an error

	messageType := int(response[0])
	
	if messageType != MessageTypeBet {
		log.Warningf("Unexpected message type received: %d, expected: %d", messageType, MessageTypeBet)
	}
	
	// decode the rest of the response from bytes to string
	decoded_response := string(response[1:])
	parts := strings.Split(decoded_response, Delimiter)
	responseType := int(parts[0][0] - '0') // We have to subtract a string 0 because a string 0 maps to int 48.
	document := parts[1]
	betAmount := parts[2]

	return responseType, document, betAmount
}

func (c *Client) ProcessResponseMultipleBet(response []byte) (int, int) {
	// This function processes the response received from the server
	// The response has the format:
	// First byte is the message type
	// The rest has the format "responseType|numberOfBets"
	// responseType is 0 if the bet was sent correctly or 1 if there was an error

	messageType := int(response[0])
	
	if messageType != MessageTypeMultipleBet {
		log.Warningf("Unexpected message type received: %d, expected: %d", messageType, MessageTypeMultipleBet)
	}
	
	decoded_response := string(response[1:])
	parts := strings.Split(decoded_response, Delimiter)
	responseType := int(parts[0][0] - '0') // We have to subtract a string 0 because a string 0 maps to int 48.
	
	numberOfBets, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Errorf("Failed to parse number of bets: %v", err)
		numberOfBets = 0
	}

	return responseType, numberOfBets
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

	responseType, document, betAmount := c.ProcessResponseSingleBet(response)

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

func (c *Client) CreateBetsPacket(bets []string) []byte {
	// This function creates a packet with a certain amount of bets
	// The packet format is as follows:
	// 4 bytes for the packet length
	// 1 byte for the packet type
	// The rest is the payload
	// The payload is formed by bets separated by a delimiter "\n"

	clientId := os.Getenv("CLI_ID")

	for i, bet := range bets {
		bets[i] = fmt.Sprintf("%s|%s", clientId, strings.ReplaceAll(bet, ",", "|"))
	}

	payload := strings.Join(bets, BetDelimiter)
	payloadLength := len(payload)

	header := make([]byte, HeaderLength)
	binary.BigEndian.PutUint32(header[:MessageTypePos], uint32(payloadLength+HeaderLength))
	header[MessageTypePos] = MessageTypeMultipleBet

	packet := append(header, []byte(payload)...)
	return packet
}

func (c *Client) ReceiveMultipleBetResponse() {
	// This function receives a response for multiple bets sent to the server
	// this response can either be a success or an error
	// It logs the result of the operation
	// The success response occurs when all the bets were sent correctly
	// The error response occurs when at least one bet was not sent correctly
	// The response has the format:
	// 2 bytes for the response length
	// then the payload with the format "responseType|numberOfBets"
	// responseType is 0 if the bets were sent correctly or 1 if there was an error

	response_length, err := c.SafeRead(2)
	if err != nil {
		log.Errorf("action: respuesta_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	response, err := c.SafeRead(int(binary.BigEndian.Uint16(response_length)))
	if err != nil {
		log.Errorf("action: respuesta_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	responseType, numberOfBets := c.ProcessResponseMultipleBet(response)

	if responseType != MessageTypeSuccess {
		log.Infof("action: respuesta_recibida | result: fail | numero_apuestas: %v",
			numberOfBets,
		)
	}
}

func (c *Client) ReceiveReadyForLotteryResponse() {
	// This function receives a response for the ready for lottery message sent to the server
	// this response can either be a success or an error
	// It logs the result of the operation
	// The success response occurs when the message was sent correctly
	// The error response occurs when the message was not sent correctly
	// The response has the format:
	// 2 bytes for the response length
	// Then 1 byte for the message type
	// Then the payload with the format "responseType"
	// responseType is 0 if the message was sent correctly or 1 if there was an error

	response_length, err := c.SafeRead(2)
	if err != nil {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	response, err := c.SafeRead(int(binary.BigEndian.Uint16(response_length)))
	if err != nil {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	
	if response[0] != MessageTypeReadyForLottery {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: unexpected message type %v, expected %v",
			c.config.ID,
			response[0],
			MessageTypeReadyForLottery,
		)
		return
	}

	responseType := int(response[1] - '0') // We have to subtract '0' because a character '0' maps to int 48.


	if responseType != MessageTypeSuccess {
		log.Infof("action: respuesta_recibida | result: fail")
	} 
}

func (c *Client) receiveLotteryFinished() (bool, int) {
	// This function receives a response for the lottery finished message sent to the server
	// this response can either be a success or an error
	// It logs the result of the operation
	// The success response occurs when the message was sent correctly
	// The error response occurs when the message was not sent correctly
	// The response has the format:
	// 2 bytes for the response length
	// Then 1 byte for the message type
	// Then the payload with the format "responseType|<winner_amount>|<winner1>|<winner2>|...|<winnerN>"
	// responseType is 0 if the message was sent and the lottery has finished
	// 1 if the message was sent but the lottery has not finished
	// 2 if the message was not sent correctly

	response_length, err := c.SafeRead(2)
	if err != nil {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return false, 0
	}

	response, err := c.SafeRead(int(binary.BigEndian.Uint16(response_length)))
	if err != nil {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return false, 0
	}

	
	if response[0] != MessageTypeAwaitingLottery {
		log.Errorf("action: confirmacion_recibida | result: fail | client_id: %v | error: unexpected message type %v, expected %v",
			c.config.ID,
			response[0],
			MessageTypeAwaitingLottery,
		)
		return false, 0
	}

	decoded_response := string(response[1:])
	parts := strings.Split(decoded_response, Delimiter)
	responseType := int(parts[0][0] - '0') // We have to subtract a string 0 because a string 0 maps to int 48.
	
	if responseType == MessageTypeSuccess {
		winnersAmount, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Errorf("Failed to parse number of winners: %v", err)
			winnersAmount = 0
		}
		return true, winnersAmount
	} else {
		return false, 0
	}
}