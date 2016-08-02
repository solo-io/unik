var express    = require('express');
var app = express();
var directory = require('serve-index');

app.use(express.static(__dirname+"/static"));
app.use(directory(__dirname+"/static"));
app.listen("8080");
