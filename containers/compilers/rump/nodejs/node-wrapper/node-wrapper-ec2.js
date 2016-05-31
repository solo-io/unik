process.chdir('/tmp');

var StdOutFixture = require('./fixture-stdout.js');
var stdOutFixture = new StdOutFixture();

//set up log server
var _log = [];
stdOutFixture.capture( function onWrite (string, encoding, fd) {
  _log.push(string);
  return true;
});
var StdErrFixture = require('./fixture-stderr');
var stdErrFixture = new StdErrFixture();
stdErrFixture.capture( function onWrite (string, encoding, fd) {
  _log.push(string);
  return true;
});

const PORT=9876;
var http = require('http');
function serveLogs(request, response){
    response.end(_log.join(""));
}
var server = http.createServer(serveLogs);
server.listen(PORT, function(){
    console.log("Log server started on: http://localhost:%s", PORT);
});

console.log("unik v0.0 boostrapping beginning udp broadcast...");
tring = require('querystring');
  var options = {
    hostname: "169.254.169.254",
    path: '/latest/user-data',
    method: 'GET',
  };
  var req = http.request(options, function(res) {
    console.log('Status: ' + res.statusCode);
    console.log('Headers: ' + JSON.stringify(res.headers));
    res.setEncoding('utf8');
    res.on('data', function (body) {
      console.log('Response: ' + body);
      env = JSON.parse(body);
      Object.keys(env).forEach(function(key) {
        var val = env[key];
        process.env[key] = val;
        console.log("Set env var: "+key+"="+val)
      });
      console.log("unik v0.0 boostrapping finished!\ncalling main");
      //CALL_NODE_MAIN_HERE
    });
  });
  req.end();
