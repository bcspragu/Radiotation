#!/usr/bin/env python
from gmusicapi import Mobileclient

from base64 import b64decode
from flask import Flask
from flask import request
from getpass import getpass

import json
import logging

app = Flask(__name__)
api = Mobileclient()
logged_in = api.login('brandon00sprague@gmail.com', b64decode(getpass() + 'wcVd5VGhlQmFzZUNhbm5vbjlzTE5w'), '36283c364b758412')
if not logged_in:
    logger.critical('could not log in successfully')
    quit()

@app.route("/search")
def search():
    return json.dumps(api.search(request.args.get('search'))['song_hits'])

@app.route("/track")
def track():
    return api.get_stream_url(request.args.get('id'))

app.run()
