//Lets require/import the HTTP module
var http = require('http');
var fs = require('fs');
var dispatcher = require('httpdispatcher');

//Lets define a port we want to listen to
const PORT=8080;

//We need a function which handles requests and send response
function handleRequest(request, response){
    try {
        //log the request on console
        console.log(request.url);
        //Disptach
        dispatcher.dispatch(request, response);
    } catch(err) {
        console.log(err);
    }
}

//Create a server
var server = http.createServer(handleRequest);

//Lets start our server
server.listen(PORT, function(){
    //Callback triggered when server is successfully listening. Hurray!
    console.log("Server listening on: http://localhost:%s", PORT);
});

//ping test
dispatcher.onGet("/ping_test", function(req, res) {
    var responseObject = {message: "pong"};
    res.end(JSON.stringify(responseObject));
});

//env test
dispatcher.onGet("/env_test", function(req, res) {
    var responseObject = {message: process.env.KEY};
    res.end(JSON.stringify(responseObject));
});

//mount_test test
dispatcher.onGet("/mount_test", function(req, res) {
    fs.readFile('/data/data.txt', 'utf8', function (err,data) {
        if (err) {
            console.log(err);
            res.end(err.toString('utf8'));
            return
        }
        var responseObject = {message: data.toString('utf8')};
        res.end(JSON.stringify(responseObject));
    });
});
