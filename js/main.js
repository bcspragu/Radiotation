var conn;

$(function() {
  setTimeout(sizeQueue, 10);
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
    var icon = $(this).find('.glyphicon');
    var form = $(this).parents('form');

    var sucFunc = function(form, room) {
      form.find('.glyphicon').removeClass('glyphicon-plus').addClass('glyphicon-ok');
      form.attr('action', '/rooms/' + room + '/remove');
    }
    if (!icon.hasClass('glyphicon-plus')) {
      sucFunc = function(form, room) {
        form.find('.glyphicon').removeClass('glyphicon-ok').addClass('glyphicon-plus');
        form.attr('action', '/rooms/' + room + '/add');
      }
    }

    $.post(form.attr('action'), form.serialize(), function(data) {
      if (data.Error) {
        alert(data.Message);
      } else {
        var room = $('.room-name').val();
        sucFunc(form, room);
        $('.queue').load('/rooms/' + room + '/queue', sizeQueue);
      }
    });
  });
});

function sizeQueue() {
  var sum = 10;
  // This is way too hacky, fix it
  $(".queue").css('width', '10000');
  var queueTracks = $(".queue-track");
  queueTracks.each(function(){sum += $(this).width()});
  if (queueTracks.length > 0) {
    $(".queue").css('width', sum);
  } else {
    $(".queue").css('width', '');
  }
}

function loadWebSockets() {
  var room = $('.room-name');
  if (window["WebSocket"] && room.length > 0) {
    var name = room.val()
    conn = new WebSocket("ws://" + host + "/rooms/" + name + "/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
      var room = $('.room-name').val();
      $('.queue').load('/rooms/' + room + '/queue', sizeQueue);
    }
  } else {
    // You ain't got WebSockets, brah
  }
}

jQuery.fn.extend({
  onAnimationEnd: function(callback) {
    $(this).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(e) {
      e.stopPropagation();
      callback();
    });
  }
});
