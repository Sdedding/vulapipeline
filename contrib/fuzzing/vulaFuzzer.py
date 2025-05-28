import atheris
from atheris.instrument_bytecode import instrument_func
import sys
import os

@atheris.instrument_func #type: ignore[misc]
def checkVerifyAgainst(hostname: str ) -> None:
	os.system('vula verify against ' + str(hostname))

@atheris.instrument_func #type: ignore[misc]
def checkVulaAlone(data: str) -> None:
	os.system('vula ' + str(data))

atheris.Setup(sys.argv, checkVerifyAgainst) #change the function name to fuzz something else
atheris.Fuzz()