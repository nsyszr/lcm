const WebSocket = require("ws");
const url = "ws://localhost:4001/api/v1/realtime-events";
const ws = new WebSocket(url);

process.stdin.resume();
process.stdin.setEncoding("utf8");

//process.stdin.on("data", function(message) {
//message = message.trim();
//ws.send(message, console.log.bind(null, "Sent : ", message));
//});

ws.on("message", function(message) {
  console.log(message);
});

ws.on("close", function(code) {
  console.log("Disconnected: " + code);
});

ws.on("error", function(error) {
  console.log("Error: " + error.code);
});
