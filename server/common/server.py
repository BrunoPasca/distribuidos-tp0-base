import socket
import logging
from common.utils import is_valid_message, Bet, store_bets, get_winners
import os
from common.constants import (
    ERROR_CODE_NO_ERRORS, 
    ERROR_CODE_INVALID_MESSAGE,
    ERROR_CODE_LOTTERY_NOT_READY,
    HEADER_LENGTH,
    MSG_LENGTH,
    MSG_TYPE_SINGLE_BET,
    MSG_TYPE_MULTIPLE_BETS,
    MSG_TYPE_READY_FOR_LOTTERY,
    MSG_TYPE_AWAITING_LOTTERY,
    DOCUMENT_POS,
    BET_AMOUNT_POS,
    RESPONSE_HEADER_LENGTH
)

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._last_bet_amount = 0
        self.agencies_waiting = set()
        self.winners = {}
        self.lottery_performed = False


    def run(self):        
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            client_sock = self.__accept_new_connection()
            self.__handle_client_connection(client_sock)

    def shutdown(self):
        self._server_socket.close()
        logging.info("action: close_server_socket | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg, msg_type = self.__read_from_socket(client_sock)
            fields, response_error_code = self.process_message(msg, msg_type)
            self.handle_response(fields, response_error_code, client_sock, msg_type)
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is loggued and returned
        """

        # Connection arrived
        c, addr = self._server_socket.accept()
        return c

    def __read_from_socket(self, sock) -> tuple[str, int]:
        """
        Read from socket

        Function reads from a socket with a specific protocol:
        - First 4 bytes represent the total length of the package
        - 5th byte represents the type of message (0 = bet, the rest are for future types of messages)
        - From 6th byte onward is the payload

        Returns a tuple with the payload and the type of message
        If the message is invalid, the payload will be an empty string and the type -1
        """
        header = self.safe_read(sock, HEADER_LENGTH)
        if not header:
            logging.error('action: receive_message | result: fail | error: short-read')
            return ("", -1)
        msg_length = int.from_bytes(header[:MSG_LENGTH], byteorder='big')
        msg_type = header[4]

        payload_length = msg_length - HEADER_LENGTH
        if payload_length > 0:
            payload = self.safe_read(sock, payload_length)
            if not payload:
                logging.error('action: receive_message | result: fail | error: short-read')
                return ("", -1)
            payload_str = payload.decode('utf-8').strip()
        else:
            payload_str = ""

        addr = sock.getpeername()        
        return (payload_str, msg_type)
    
    def safe_read(self, sock, length):
        data = b''
        while len(data) < length:
            packet = sock.recv(length - len(data))
            if not packet: # This means the connection was closed and a short-read occurred
                return None
            data += packet
        return data
    
    def safe_send(self, sock, data):
        total_sent = 0
        while total_sent < len(data):
            sent = sock.send(data[total_sent:])
            if sent == 0: # This means the connection was closed and a short-write occurred
                return False
            total_sent += sent
        return True
    
    def process_message(self, message, msg_type):
        """
        Process message

        Function processes message and returns a response
        """
        if msg_type == MSG_TYPE_SINGLE_BET:
            return self.process_message_single_bet(message)
        elif msg_type == MSG_TYPE_MULTIPLE_BETS:
            return self.process_message_multiple_bets(message)
        elif msg_type == MSG_TYPE_READY_FOR_LOTTERY:
            return self.process_message_ready_for_lottery(message)
        elif msg_type == MSG_TYPE_AWAITING_LOTTERY:
            return self.process_message_awaiting_lottery(message)
        return ((), ERROR_CODE_INVALID_MESSAGE)

    def process_message_single_bet(self, message):
        """
        This function processes a single bet message and stores it in the bets file
        """
        fields, status = is_valid_message(message)
        if status == ERROR_CODE_INVALID_MESSAGE:
            logging.error(f"action: receive_message | result: fail | error: invalid message")
            return ((), ERROR_CODE_INVALID_MESSAGE)
        store_bets([Bet(*fields)])
        agency, name, last_name, document, birthdate, number = fields # Some fields might be useful in the future
        self.agencies.add(agency)
        logging.info(f"action: apuesta_almacenada | result: success | dni: {document} | numero: {number}")
        return (fields, ERROR_CODE_NO_ERRORS)
    
    def process_message_multiple_bets(self, message):
        """
        This function processes multiple bets at a time
        If all bets are valid, they are stored in the bets file
        If at least one bet is invalid, no bets are stored

        Bets are delimited by a newline character "\n"

        Inside each bet, fields are separated by a "|"
        """
        bets = message.split("\n")
        bet_amount = len(bets)
        self._last_bet_amount = bet_amount
        valid_bets = []
        for bet in bets:
            fields, status = is_valid_message(bet)
            if status == ERROR_CODE_INVALID_MESSAGE:
                logging.error(f"action: apuesta_recibida | result: fail | cantidad: {bet_amount}")
                return ((), ERROR_CODE_INVALID_MESSAGE)
            valid_bets.append(Bet(*fields))
        store_bets(valid_bets)
        logging.info(f"action: apuesta_recibida | result: success | cantidad: {bet_amount}")
        return (fields, ERROR_CODE_NO_ERRORS)
    
    def process_message_ready_for_lottery(self, message):
        """
        This function processes the message that indicates an agency has finished
        sending bets and is ready for the lottery
        """
        if not message.isdigit():
            return ((), ERROR_CODE_INVALID_MESSAGE)
        agency_id = int(message)
        self.agencies_waiting.add(agency_id)

        client_amount = int(os.getenv('CLIENT_AMOUNT', 0))

        if len(self.agencies_waiting) == client_amount:
            logging.info(f"action: sorteo | result: success")
            winners = get_winners()
            self.winners = winners
            self.lottery_performed = True
            return ((), ERROR_CODE_NO_ERRORS)
        return ((), ERROR_CODE_NO_ERRORS)

    def process_message_awaiting_lottery(self, message):
        """
        This function processes the message that indicates a client is waiting for the lottery
        If we have winners then that means that the lottery already happened.
        We return whether there are winners or not
        """
        if not message.isdigit():
            return ((), ERROR_CODE_INVALID_MESSAGE)
        agency_id = int(message)
        
        if not self.lottery_performed:
            return ((), ERROR_CODE_LOTTERY_NOT_READY)
            
        if agency_id not in self.winners:
            return ([], ERROR_CODE_NO_ERRORS)
        
        winners = self.winners[agency_id]
        return (winners, ERROR_CODE_NO_ERRORS)

    
    def handle_response(self, fields, response_error_code, sock, msg_type):
        """
        This functions determines which functions processes the response
        depending on the message type
        """
        if msg_type == MSG_TYPE_SINGLE_BET:
            self.handle_bet_response(fields, response_error_code, sock)
        elif msg_type == MSG_TYPE_MULTIPLE_BETS:
            self.handle_multiple_bets_response(fields, response_error_code, sock)
        elif msg_type == MSG_TYPE_READY_FOR_LOTTERY:
            self.handle_ready_for_lottery_response(fields, response_error_code, sock)
        elif msg_type == MSG_TYPE_AWAITING_LOTTERY:
            self.handle_awaiting_lottery_response(fields, response_error_code, sock)
        pass 

    def handle_bet_response(self, fields, response_error_code, sock):
        """
        This function sends a response to the client depending on the
        whether the bet was received correctly or not.

        If the bet was received correctly, the response will be:
        2 bytes: length of the response
        1 byte: message type
        Then the payload will be: 0|<document>|<number>
        If the bet was not received correctly, the response will be:
        2 bytes: length of the response
        1 byte: message type
        Then the payload will be: 1|<document>|<number>
        """

        if response_error_code == ERROR_CODE_NO_ERRORS:
            response_payload = f"{ERROR_CODE_NO_ERRORS}|{fields[DOCUMENT_POS]}|{fields[BET_AMOUNT_POS]}"
        else:
            response_payload = f"{ERROR_CODE_INVALID_MESSAGE}"
        
        response_bytes = response_payload.encode('utf-8')
        response_with_type = bytes([MSG_TYPE_SINGLE_BET]) + response_bytes
        
        response_length = len(response_with_type).to_bytes(RESPONSE_HEADER_LENGTH, byteorder='big')
        response = response_length + response_with_type
        
        self.safe_send(sock, response)
    
    def handle_multiple_bets_response(self, fields, response_error_code, sock):
        """
        This function sends a response to the client depending on the
        whether the bets were received correctly or not.

        If the bets were received correctly, the response will be:
        2 bytes: length of the response
        1 byte: message type
        Then the payload will be: 0|<number_of_bets>
        If the bets were not received correctly, the response will be:
        2 bytes: length of the response
        1 byte: message type
        Then the payload will be: 1|<number_of_bets>
        """

        response_payload = f"{response_error_code}|{self._last_bet_amount}"
        
        response_bytes = response_payload.encode('utf-8')
        response_with_type = bytes([MSG_TYPE_MULTIPLE_BETS]) + response_bytes
        
        response_length = len(response_with_type).to_bytes(RESPONSE_HEADER_LENGTH, byteorder='big')
        response = response_length + response_with_type
        
        self.safe_send(sock, response)

    def handle_ready_for_lottery_response(self, fields, response_error_code, sock):
        """
        This function sends a response to the client depending on the
        whether the agency is ready for the lottery or not.

        If the agency is ready for the lottery, the response will be:
        2 byte the length of the response
        1 byte: message type
        Then the payload will be: 0 or 1 depending on the response_error_code
        """

        response_payload = f"{response_error_code}"
        
        response_bytes = response_payload.encode('utf-8')
        response_with_type = bytes([MSG_TYPE_READY_FOR_LOTTERY]) + response_bytes
        
        response_length = len(response_with_type).to_bytes(RESPONSE_HEADER_LENGTH, byteorder='big')
        response = response_length + response_with_type

        self.safe_send(sock, response)

    def handle_awaiting_lottery_response(self, fields, response_error_code, sock):
        """
        This function sends a response to the client letting them know
        if the lottery happened or not.
        If the lottery happened, the response will be:
        2 byte the length of the response
        1 byte: message type
        Then the payload will be: 0|<number_of_winners>
        If the lottery did not happen, the response will be:
        2 byte the length of the response
        1 byte: message type
        Then the payload will be: 1 or 2 (unexpected agency or lottery not ready)
        """
        if response_error_code == ERROR_CODE_NO_ERRORS:
            response_payload = f"{ERROR_CODE_NO_ERRORS}|{len(fields)}"
        else:
            response_payload = f"{response_error_code}"
        
        response_bytes = response_payload.encode('utf-8')
        response_with_type = bytes([MSG_TYPE_AWAITING_LOTTERY]) + response_bytes
        
        response_length = len(response_with_type).to_bytes(RESPONSE_HEADER_LENGTH, byteorder='big')
        response = response_length + response_with_type

        self.safe_send(sock, response)
