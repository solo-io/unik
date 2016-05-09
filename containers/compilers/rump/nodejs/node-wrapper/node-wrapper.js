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

console.log('UDP Server starting');
const HOST = '0.0.0.0';
var dgram = require('dgram');
var udpServer = dgram.createSocket('udp4');
udpServer.on('listening', function () {
    var address = server.address();
    console.log('UDP Server listening on ' + address.address + ":" + address.port);
});
var listenerIp = "";
udpServer.on('message', function (message, remote) {
    message = message.toString().replace(/\0/g, '');
    console.log("udp listener received: "+remote.address + ':' + remote.port +' - ' + message + " length: "+message.length);
    if (message.indexOf("unik") > -1) {
      listenerIp = message.split(":")[1]
      registerWithListener(listenerIp)
      udpServer.close();
    }
});
udpServer.bind(PORT, HOST);

console.log("unik v0.0 boostrapping beginning udp broadcast...");
function registerWithListener(listenerIp) {
  require('macaddress').one(function (err, mac) {
    console.log("Mac address for this host: %s", mac);
    var querystring = require('querystring');
    var options = {
      hostname: listenerIp,
      port: 3000,
      path: '/register?mac_address='+mac,
      method: 'POST',
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
  });
}
