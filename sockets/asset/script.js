var ifAuth = false;
var socket = io();
var d = new Date();
var usersConnect = [];

var text = "";
var name = "";

var id = d.getTime().toString();
var nameId = (d.getTime() * 2).toString();

clientCursorColor = "#" + Math.floor(Math.random() * 700000 + 100000);

$('#btn').click(function () {
  ifAuth = true;
  name = $('#nameInput').val() + "";
  socket.emit('user', name, clientCursorColor, nameId, id);

  $('span').show();
  $('input').hide();

  $("#field").mousemove(function (event) {
    var coordsX = event.pageX + "";
    var coordsY = event.pageY + "";

    socket.emit('coords', coordsX, coordsY);
  });
});

socket.on('get info', function () {

  socket.emit('user', name, clientCursorColor, nameId, id);
})

socket.on('append user', function (name, color, idName) {
  if (searchIDInList(idName)) {
    $('#usersConnect').append($("<div>", {
      "class": "names",
      "id": idName
    }).text(name).css('color', color));
  }
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

socket.on('close', function (cursor, name) {

  ifAuth = false;
  $('#' + cursor).remove();
  $('#' + name).remove();
});

function searchIDInList(id) {
  if (!usersConnect.includes(id)) {
    usersConnect.push(id);
    return true;
  }
  return false;
}