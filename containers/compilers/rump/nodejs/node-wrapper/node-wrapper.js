var StdOutFixture = require('./fixture-stdout.js');
var stdOutFixture = new StdOutFixture();

var _log = [];

stdOutFixture.capture( function onWrite (string, encoding, fd) {
  _log.push(string);
  return true;
});

var StdErrFixture = require('./fixture-stderr');
var stdErrFixture = new StdErrFixture();

// Capture a write to stderr
stdErrFixture.capture( function onWrite (string, encoding, fd) {
  _log.push(string);
  return true;
});

console.log(_log.join(""))
