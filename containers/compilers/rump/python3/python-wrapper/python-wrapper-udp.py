import http.client
import json
from io import StringIO
import sys
import threading
import socket
from http.server import BaseHTTPRequestHandler, HTTPServer
from uuid import getnode as get_mac

###From https://github.com/bmc/grizzled-python/blob/bf9998bd0f6497d1e368610f439f9085d019bf76/grizzled/io/__init__.py
# ---------------------------------------------------------------------------
# Imports
# ---------------------------------------------------------------------------

import os
import zipfile

class MultiWriter(object):
    """
    Wraps multiple file-like objects so that they all may be written at once.
    For example, the following code arranges to have anything written to
    ``sys.stdout`` go to ``sys.stdout`` and to a temporary file:
    .. python::
        import sys
        from grizzled.io import MultiWriter
        sys.stdout = MultiWriter(sys.__stdout__, open('/tmp/log', 'w'))
    """
    def __init__(self, *args):
        """
        Create a new ``MultiWriter`` object to wrap one or more file-like
        objects.
        :Parameters:
            args : iterable
                One or more file-like objects to wrap
        """
        self.__files = args

    def write(self, buf):
        """
        Write the specified buffer to the wrapped files.
        :Parameters:
            buf : str or bytes
                buffer to write
        """
        for f in self.__files:
            f.write(buf)

    def flush(self):
        """
        Force a flush.
        """
        for f in self.__files:
            f.flush()

    def close(self):
        """
        Close all contained files.
        """
        for f in self.__files:
            f.close()

logsbuf = StringIO()

class Capturing(list):
    def __enter__(self):
        self._stdout = sys.stdout
        self._stderr = sys.stderr
        sys.stdout = self._stringioout = MultiWriter(logsbuf, self._stdout)
        sys.stderr = self._stringioerr = MultiWriter(logsbuf, self._stderr)
        return self
    def __exit__(self, *args):
        self.extend(logsbuf.getvalue().splitlines())
        self.extend(logsbuf.getvalue().splitlines())
        sys.stdout = self._stdout
        sys.stderr = self._stderr

CONST_PORT=9876

# HTTPRequestHandler class
class LogServer(BaseHTTPRequestHandler):
  # GET
  def do_GET(self):
        # Send response status code
        self.send_response(200)

        # Send headers
        self.send_header('Content-type','text/html')
        self.end_headers()

        # Send message back to client
        message = logsbuf.getvalue()
        # Write content as utf-8 data
        self.wfile.write(bytes(message, "utf8"))
        return

def run():
  print('starting server on port', CONST_PORT)
  server_address = ('127.0.0.1', CONST_PORT)
  httpd = HTTPServer(server_address, LogServer)
  print('running server...')
  httpd.serve_forever()

def registerWithListener(listenerip):
    connection = http.client.HTTPConnection(listenerip, 3000)
    mac = get_mac()
    mac = ':'.join(("%012X" % mac)[i:i+2] for i in range(0, 12, 2)).lower()
    connection.request('POST', '/register?mac_address='+mac)
    print("sending POST",listenerip,':3000','/register?mac_address='+mac)
    response = connection.getresponse()
    env = json.loads(response.read().decode())
    print("received", env)
    for key in env:
        os.environ[key] = env[key]

with Capturing():
    t = threading.Thread(target = run)
    t.daemon = True
    t.start()

    UDP_IP = "0.0.0.0"
    UDP_PORT = 9876

    sock = socket.socket(socket.AF_INET, # Internet
                         socket.SOCK_DGRAM) # UDP
    sock.bind((UDP_IP, UDP_PORT))

    while True:
        data, addr = sock.recvfrom(1024)
        print ("received message:", data)
        if "unik" in data.decode("utf-8"):
            listenerip = data.decode("utf-8").split(':')[1]
            registerWithListener(listenerip)
            break

    import main.py ##NAME OF SCRIPT GOES HERE
