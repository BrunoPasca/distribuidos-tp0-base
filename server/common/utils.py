import csv
import datetime
import time

from common.constants import (
    DELIMITER,
    STORAGE_FILEPATH,
    LOTTERY_WINNER_NUMBER,
    MAX_NAME_LENGTH,
    MAX_AMOUNT
    )

""" A lottery bet registry. """
class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])

""""
Validates the input of a message is considered valid.
"""
def is_valid_message(message: str) -> bool:
    """
    A valid message must have the following format:
    agency|name|last_name|document|birthdate|number|
    where:
    - name, last_name: only alphabetical characters
    - document: 8 digits
    - birthdate: 'YYYY-MM-DD'
    - number: a positive integer

    Returns a tuple with the fields of the message and a status code:
    - 0: valid message
    - 1: invalid
    """
    fields = message.split(DELIMITER)
    if len(fields) != 6:
        return ([], 1)
    agency, name, last_name, document, birthdate, number = fields
    if not agency.isdigit() or int(agency) < 0:
        return ([], 1)
    if not _validate_name(name) or not _validate_name(last_name):
        return ([], 1)
    if not document.isdigit() or len(document) != 8:
        return ([], 1)
    if not _validate_date(birthdate):
        return ([], 1)
    if not number.isdigit() or int(number) < 0 or int(number) > MAX_AMOUNT:
        return ([], 1)
    return (fields, 0)
    
def _validate_name(name: str) -> bool:
    if len(name) == 0 or len(name) > MAX_NAME_LENGTH:
        return False
    for char in name:
        if not char.isalpha() and not char.isspace() and char not in ['.', '-', "'"]:
            return False
    return True

def _validate_date(date: str) -> bool:
    try:
        datetime.datetime.strptime(date, '%Y-%m-%d')
        return True
    except ValueError:
        return False
    
def get_winners() -> dict:
    """
    Returns a dictionary with the winners of the lottery for each agency.
    """
    winners = {}
    for bet in load_bets():
        if has_won(bet):
            winners[bet.agency] = winners.get(bet.agency, [])
            winners[bet.agency].append(bet.document)
    return winners