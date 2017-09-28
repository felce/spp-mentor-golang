var socket = io();
var d = new Date();

var id = d.getTime().toString();
clientCursorColor = "#" + Math.floor(Math.random() * 700000 + 100000);

$("#field").mousemove(function (event) {
  var coordsX = event.pageX + "";
  var coordsY = event.pageY + "";

  socket.emit('coords', coordsX, coordsY, clientCursorColor, id);
});

socket.on('new coords', function (coordsX, coordsY, color, index) {

  $('#' + index).remove();
  var styles = {
    backgroundColor: color,
    top: coordsY * 1,
    left: coordsX * 1
  };
  $('#field').append($("<div>", {
    "class": "anotherUsersCursor",
    "id": index
  }).css(styles));
});

socket.on('close', function (index) {
  console.log(index);
  $('#' + index).remove();
});