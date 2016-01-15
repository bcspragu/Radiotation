var conn;

$(function() {
  loadWebSockets();

  $('.room-form').submit(function(e) {
    var form = $(this);
    var room = form.find('.room-name').val();
    form.attr('action', '/rooms/' + room);
    form.find('.room-name').remove();
  });

  $('.search-form').submit(function(e) {
    e.preventDefault();
    var form = $(this);
    var res = $('.results');
    res.load(form.attr('action'), form.serialize(), function() {
      res.removeClass('hide').addClass('fadeIn animated');
    });
  });

  $('.results').on('click', '.add a', function(e) {
    e.preventDefault();

    var form = $(this).parents('form');
    $.post(form.attr('action'), form.serialize(), function(data) {
      if (data.Error) {
        alert(data.Message);
      } else {
        $('.queue').load('/rooms/' + room + '/queue');
      }
    });
  });
});

function loadWebSockets() {
  if (window["WebSocket"] && room !== '') {
    conn = new WebSocket("ws://" + host + "/rooms/" + room + "/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
      $('.queue').load('/rooms/' + room + '/queue');
    }
  } else {
    // You ain't got WebSockets, brah
  }
}
