from io import StringIO
import sys
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer

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

CONST_PORT=9967

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
  server_address = ('0.0.0.0', CONST_PORT)
  httpd = HTTPServer(server_address, LogServer)
  print('running server...')
  httpd.serve_forever()

with Capturing():
    os.chdir("/bootpart")
    t = threading.Thread(target = run)
    t.daemon = True
    t.start()

    import main.py ##NAME OF SCRIPT GOES HERE
