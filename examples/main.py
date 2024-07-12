# please see the documentation on how to build chew for use with python

import ctypes

chew_lib = ctypes.CDLL('./chew.so')

chew_lib.Process.argtypes = [ctypes.c_char_p]
chew_lib.Process.restype = ctypes.c_char_p

urls = "https://example.com"
result = chew_lib.Process(urls.encode('utf-8'))

print(result.decode('utf-8'))
