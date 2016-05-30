from bottle import *

@get('/ping_test')
def ping_test():
    return {"message": "pong"}

run(host='localhost', port=8080)
