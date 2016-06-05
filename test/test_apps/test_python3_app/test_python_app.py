from bottle import *
import os

@get('/ping_test')
def ping_test():
    return {"message": "pong"}

@get('/env_test')
def env_test():
    return {"message": os.environ['KEY']}

@get('/mount_test')
def mount_test():
    with open("/data/data.txt", "rt") as in_file:
        text = in_file.read()
        return {"message": text}

run(host='0.0.0.0', port=8080)
