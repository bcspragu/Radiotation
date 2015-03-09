var conn;

$(function() {
  $('.first').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function () {
    $('.middle').addClass('animated fadeIn').removeClass('invisible');
  });
  setTimeout(function () {
    $('.first').addClass('move');
  }, 1000);
  loadWebSockets();
});

function loadWebSockets() {
  if (window["WebSocket"]) {
    conn = new WebSocket("ws://" + host + "/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
      // Remove song from queue
    }
  } else {
    // You ain't got WebSockets, brah
  }
}
