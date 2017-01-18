var express = require('express');
var app = express();

var PORT = 8001;

console.log('Starting NodeJS example application...');

app.get('/', function (req, res) {
  res.send('Hello World!');
});

if(process.env.PORT){
	PORT = parseInt(process.env.PORT);
	console.log("PORT = " + PORT);
}

app.listen(PORT, function () {
  console.log('Example app listening on port ' + PORT + '!');
  console.log('NodeJS Version: ' + process.version);
});