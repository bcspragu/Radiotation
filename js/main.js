var conn;

$(function() {
  loadWebSockets();

  $('.search-form').submit(function(e) {
    e.preventDefault();
    reloadResults();
  });

  $('body').on('click', '.update', function(e) {
    e.preventDefault();

    var form = $(this).find('form');
    var icon = $(this).find('.fa');
    $.post(form.attr('action'), form.serialize(), function(data) {
      if (data.Error) {
        alert(data.Message);
      } else {
        icon.toggleClass('fa-check');
        icon.toggleClass('fa-plus');
        $('.queue').load('/rooms/' + room + '/queue');
        reloadResults();
      }
    });
  });
});

function reloadResults() {
  var form = $('.search-form');
  if (form.find('input[name="search"]').val() !== '') {
    var res = $('.results');
    res.load(form.attr('action'), form.serialize());
  }
}

function loadWebSockets() {
  if (window["WebSocket"] && typeof(room) !== 'undefined' && room != '') {
    conn = new WebSocket("ws://" + host + "/rooms/" + room + "/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
      if (evt.data === "pop") {
        $('.queue').load('/rooms/' + room + '/queue');
        reloadResults();
        return;
      }

      if (evt.data === "playing") {
        $('.now-playing').load('/rooms/' + room + '/now');
        return;
      }
    }
  } else {
    // You ain't got WebSockets, brah
  }
}
