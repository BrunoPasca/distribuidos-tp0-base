""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574
# Other Constants
MSG_LENGTH = 1024
ERROR_CODE_NO_ERRORS = 0
ERROR_CODE_INVALID_MESSAGE = 1
HEADER_LENGTH = 5
MSG_LENGTH = 4 # 4 bytes for the msg length
MSG_TYPE_SINGLE_BET = 0
MSG_TYPE_MULTIPLE_BETS = 1
DELIMITER = '|' # Delimiter used to separate fields in a message
DOCUMENT_POS = 3
BET_AMOUNT_POS = 5
RESPONSE_HEADER_LENGTH = 2
MAX_RETRIES = 3

MAX_NAME_LENGTH = 50
MAX_AMOUNT = 999999