var conn;

$(function() {
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
    res.load('/search', form.serialize(), function() {
      res.removeClass('hide').addClass('fadeIn animated');
    });
  });

  $('.results').on('click', '.add a', function(e) {
    e.preventDefault();
    var icon = $(this).find('.glyphicon');
    if (icon.hasClass('glyphicon-plus')) {
      var form = $(this).parents('form');
      $.post("/add", form.serialize(), function(data) {
        if (data.Error) {
          alert(data.Message);
        } else {
          form.find('.glyphicon').removeClass('glyphicon-plus').addClass('glyphicon-ok');
          var queue = $('.queue');
          queue.load('/queue', sizeQueue);
        }
      });
    } else {
      var form = $(this).parents('form');
      $.post("/remove", form.serialize(), function(data) {
        if (data.Error) {
          alert(data.Message);
        } else {
          form.find('.glyphicon').removeClass('glyphicon-ok').addClass('glyphicon-plus');
          var queue = $('.queue');
          queue.load('/queue', sizeQueue);
        }
      });

    }
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

jQuery.fn.extend({
  onAnimationEnd: function(callback) {
    $(this).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(e) {
      e.stopPropagation();
      callback();
    });
  }
});
