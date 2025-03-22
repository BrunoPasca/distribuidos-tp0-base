import socket
import logging
from utils import is_valid_message, Bet, store_bets

MSG_LENGTH = 1024

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
        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
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
            # TODO: Modify the receive to avoid short-reads
            msg = self.__read_from_socket(client_sock)
            self.process_message(msg, client_sock.getpeername()[0])
            
            # TODO: Modify the send to avoid short-writes
            client_sock.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
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
        header = self.safe_read(sock, 5)
        if not header:
            return ("", -1)
            
        msg_length = int.from_bytes(header[:4], byteorder='big')
        msg_type = header[4]
        
        payload_length = msg_length - 5
        if payload_length > 0:
            payload = self.safe_read(sock, payload_length)
            if not payload:
                return ("", -1)
            payload_str = payload.decode('utf-8').strip()
        else:
            payload_str = ""
            
        addr = sock.getpeername()
        logging.info(f'action: receive_message | result: success | ip: {addr[0]} | type: {msg_type} | payload: {payload_str}')
        
        return (payload_str, msg_type)

    def process_message(self, message, sender):
        """
        Process message

        Function processes message and returns a response
        """
        fields, status = is_valid_message(message)
        if status == 1:
            logging.error(f"action: receive_message | result: fail | error: invalid message | ip: {sender}")
            return 1
        logging.info(f'action: receive_message | result: success | ip: {sender} | msg: {message}')
        store_bets([Bet(*fields)])
        agency, name, last_name, document, birthdate, number = fields # Some fields might be useful in the future
        logging.info(f"action: apuesta_almacenada | result: success | dni: ${document} | numero: ${number}")
        return 0
    
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
    