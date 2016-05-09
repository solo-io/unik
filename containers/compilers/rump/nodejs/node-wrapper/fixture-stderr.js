/**
 * Module Dependencies
 */
var util = require('util');
var Stream = require('stream');


/**
 * Test fixture which globally intercepts writes
 * to stderr.
 *
 * Based on: https://gist.github.com/pguillory/729616
 *
 * @option {Stream}				[stream to intercept to-- defaults to stderr]
 *
 * @return {Function}          [an instance of the fixture]
 */

var StdErrFixture = function ( options ) {

	// Options
	if ( typeof options !== 'object' ) options = {};
	if ( options instanceof Stream ) options = { stream: options };
	var stream = options.stream || process.stderr;

	// Replace stderr
	var _intercept = function (callback) {
		var original_stderr_write = stream.write;

		stream.write = (function (write) {
			return function (string, encoding, fd) {
				var interceptorReturnedFalse = false === callback(string, encoding, fd);
				if (interceptorReturnedFalse) return;
				else write.apply(stream, arguments);
			};
		})(stream.write);

		return function _revert () {
			stream.write = original_stderr_write;
		};
	};

	// Revert to the original stderr
	var _release;


	/**
	 * [Capture writes sent to stderr]
	 * @param  {[type]} interceptFn [run each time a write is intercepted]
	 */
	this.capture = function (interceptFn) {

		// Default interceptFn
		interceptFn = interceptFn || function (string, encoding, fd) {
			util.debug('(intercepted a write to stderr) ::\n' + util.inspect(string));
		};

		// Save private `release` method for use later.
		_release = _intercept(interceptFn);
	};

	/**
	 * Stop capturing writes to stderr
	 */
	this.release = function () {
		_release();
	};
};



// Export the constructor
module.exports = StdErrFixture;
