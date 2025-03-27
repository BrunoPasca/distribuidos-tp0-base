import socket
import logging
from common.utils import is_valid_message, Bet, store_bets
from common.constants import (
    ERROR_CODE_NO_ERRORS, 
    ERROR_CODE_INVALID_MESSAGE,
    HEADER_LENGTH,
    MSG_LENGTH,
    MSG_TYPE_BET,
    DOCUMENT_POS,
    BET_AMOUNT_POS,
    RESPONSE_HEADER_LENGTH,
    MAX_RETRIES
)

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

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
            fields, response_error_code = self.process_message(msg, client_sock.getpeername()[0])
            self.handle_response(fields, response_error_code, client_sock, msg_type)
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
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

    def process_message(self, message, sender):
        """
        Process message

        Function processes message and returns a response
        """
        fields, status = is_valid_message(message)
        if status == ERROR_CODE_INVALID_MESSAGE:
            logging.error(f"action: receive_message | result: fail | error: invalid message | ip: {sender}")
            return ((), ERROR_CODE_INVALID_MESSAGE)
        store_bets([Bet(*fields)])
        agency, name, last_name, document, birthdate, number = fields # Some fields might be useful in the future
        logging.info(f"action: apuesta_almacenada | result: success | dni: {document} | numero: {number}")
        return (fields, ERROR_CODE_NO_ERRORS)
    
    def safe_read(self, sock, length, max_retries=MAX_RETRIES):
        data = b''
        retries = 0
        while len(data) < length:
            try:
                packet = sock.recv(length - len(data))
                if not packet:  # This means the connection was closed and a short-read occurred
                    retries += 1
                    if retries >= max_retries:
                        return None
                    continue
                data += packet
            except:
                retries += 1
                if retries >= max_retries:
                    return None
        return data
    
    def safe_send(self, sock, data):
        total_sent = 0
        while total_sent < len(data):
            sent = sock.send(data[total_sent:])
            if sent == 0: # This means the connection was closed and a short-write occurred
                return False
            total_sent += sent
        return True
    
    def handle_response(self, fields, response_error_code, sock, msg_type):
        """
        This functions determines which functions processes the response
        depending on the message type
        """
        if msg_type == MSG_TYPE_BET:
            return self.handle_bet_response(fields, response_error_code, sock)
        return False # For now we should only be getting bets

    def handle_bet_response(self, fields, response_error_code, sock):
        """
        This function sends a response to the client depending on the
        whether the bet was received correctly or not.

        If the bet was received correctly, the response will be:
        2 bytes: length of the response
        Then the payload will be: 0|<document>|<number>
        If the bet was not received correctly, the response will be:
        2 bytes: length of the response
        Then the payload will be: 1|<document>|<number>
        """

        if response_error_code == ERROR_CODE_NO_ERRORS:
            response = f"{ERROR_CODE_NO_ERRORS}|{fields[DOCUMENT_POS]}|{fields[BET_AMOUNT_POS]}"
        else:
            response = f"{ERROR_CODE_INVALID_MESSAGE}"
        
        response = response.encode('utf-8')
        response_length = len(response).to_bytes(RESPONSE_HEADER_LENGTH, byteorder='big')
        response = response_length + response
        return self.safe_send(sock, response)
        