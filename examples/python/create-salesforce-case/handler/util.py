import sys, os

# GLOBAL_VARS
LOG_ENABLED=False

if(os.getenv("write_debug")):
    sys.stderr.write(f"\033[93mWARNING!! DEBUG has been enabled for this function. Sensitive information could be printed to sysout\033[0m\n")
    LOG_ENABLED=True

class Logger:
    """
    Logger Util class made for logging to the console in FaaS setup
    """
    def __init__(self) -> None:
        self.HEADER = '\033[95m'
        self.OKBLUE = '\033[94m'
        self.OKGREEN = '\033[92m'
        self.WARNING = '\033[93m'
        self.FAIL = '\033[91m'
        self.ENDC = '\033[0m'
        self.BOLD = '\033[1m'
        self.UNDERLINE = '\033[4m'
    
    def log(self, type,k,v=''):
        """
        Log function that is configurable to provide colored messages

        Arguments:
            type {str} -- takes one of the following types 
            - TITLE, BOLD, UNDERLINE, ERROR, WARN, SUCCESS, INFO
            k {str} -- the message that you want printed

        Keyword Arguments:
            v {str} -- any value that you want printed after the color formatted message (default: {''})
        """
        if LOG_ENABLED:
            if type == "TITLE":
                sys.stderr.write(f"{self.HEADER}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "BOLD":
                sys.stderr.write(f"{self.BOLD}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "UNDERLINE":
                sys.stderr.write(f"{self.UNDERLINE}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "ERROR":
                sys.stderr.write(f"{self.FAIL}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "WARN":
                sys.stderr.write(f"{self.WARNING}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "SUCCESS":
                sys.stderr.write(f"{self.OKGREEN}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            elif type == "INFO":
                sys.stderr.write(f"{self.OKBLUE}{k}{self.ENDC}{v} \n") #Syserr only get logged on the console logs
            else:
                sys.stderr.write(f"{k}{v}\n") #Syserr only get logged on the console logs


class FaaSResponse:
    """
    FaaSResponse is a helper class to construct a properly formatted message returned by this function.
    By default, OpenFaaS will marshal this response message as JSON.
    """    
    def __init__(self, status, message):
        """
        Arguments:
            status {str} -- the response status code
            message {str} -- the response message
        """    
        self.status=status
        self.message=message