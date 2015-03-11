var conn;

$(function() {
  startAnimation();
  loadWebSockets();
});

function startAnimation() {
  // Move R over
  setTimeout(function () {
    $('.first').addClass('move');
  }, 500);

  // Fade in "adi"
  $('.first').onAnimationEnd(function () {
    $('.middle').addClass('animated fadeIn').removeClass('invisible');
  });

  // Fade out everything
  $('.middle').onAnimationEnd(function () {
    setTimeout(function() {
      $('.splash').addClass('fadeOut animated');
    }, 500);
  });
  
  // Fade in content
  $('.splash').onAnimationEnd(function () {
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
  onAnimationEnd: function (callback) {
    $(this).one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function (e) {
      e.stopPropagation();
      callback();
    });
  }
});
