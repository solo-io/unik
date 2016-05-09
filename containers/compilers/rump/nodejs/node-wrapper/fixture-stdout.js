/**
 * Module Dependencies
 */
var util = require('util');
var Stream = require('stream');


/**
 * Test fixture which globally intercepts writes
 * to stdout.
 *
 * Based on: https://gist.github.com/pguillory/729616
 *
 * @option {Stream}				[stream to intercept to-- defaults to stdout]
 *
 * @return {Function}          [an instance of the fixture]
 */

var StdOutFixture = function ( options ) {

	// Options
	if ( typeof options !== 'object' ) options = {};
	if ( options instanceof Stream ) options = { stream: options };
	var stream = options.stream || process.stdout;

	// Replace stdout
	var _intercept = function (callback) {
		var original_stdout_write = stream.write;

		stream.write = (function (write) {
			return function (string, encoding, fd) {
				var interceptorReturnedFalse = false === callback(string, encoding, fd);
				if (interceptorReturnedFalse) return;
				else write.apply(stream, arguments);
			};
		})(stream.write);

		return function _revert () {
			stream.write = original_stdout_write;
		};
	};

	// Revert to the original stdout
	var _release;


	/**
	 * [Capture writes sent to stdout]
	 * @param  {[type]} interceptFn [run each time a write is intercepted]
	 */
	this.capture = function (interceptFn) {

		// Default interceptFn
		interceptFn = interceptFn || function (string, encoding, fd) {
			util.debug('(intercepted a write to stdout) ::\n' + util.inspect(string));
		};

		// Save private `release` method for use later.
		_release = _intercept(interceptFn);
	};

	/**
	 * Stop capturing writes to stdout
	 */
	this.release = function () {
		_release();
	};
};



// Export the constructor
module.exports = StdOutFixture;
