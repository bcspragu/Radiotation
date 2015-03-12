var conn;

$(function() {
  startAnimation();
  loadWebSockets();

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
          queue.load('/queue');
        }
      });
    }
  });
});

function startAnimation() {
  // Move R over
  setTimeout(function() {
    $('.first').addClass('move');
  }, 250);

  // Fade in "adi"
  $('.first').onAnimationEnd(function() {
    $('.middle').addClass('animated fadeIn').removeClass('invisible');
  });

  // Fade out everything
  $('.middle').onAnimationEnd(function() {
    setTimeout(function() {
      $('.splash').addClass('fadeOut animated');
    }, 250);
  });
  
  // Fade in content
  $('.splash').onAnimationEnd(function() {
    $(this).remove();
    $('.content').removeClass('hide').addClass('fadeIn animated');
  });

}

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

jQuery.fn.extend({
  onAnimationEnd: function(callback) {
    $(this).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function(e) {
      e.stopPropagation();
      callback();
    });
  }
});
