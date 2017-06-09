var conn;

$(function() {
  renderButton();
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
        reloadQueue(true);
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

function reloadQueue(scroll) {
  $('.queue').load('/rooms/' + room + '/queue', function() {
    if (scroll) {
      var q = $('.queue').get(0);
      q.scrollLeft = q.scrollWidth;
    }
  });
}

function loadWebSockets() {
  if (window["WebSocket"] && typeof(room) !== 'undefined' && room != '') {
    conn = new WebSocket("wss://" + host + "/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
      if (evt.data === "pop") {
        reloadQueue(false);
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

function onSignIn(googleUser) {
  var idToken = googleUser.getAuthResponse().id_token;
  verifyToken(idToken);
}

function onFailure(err) {
  console.log(err);
}

function verifyToken(token, first, last) {
  $.ajax({
    url: '/verifyToken',
    type: 'post',
    data: {
      token: token,
    },
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    success: function(data) {
       console.log(data);
    }
  })
}

function renderButton() {
  gapi.signin2.render('g-signin', {
    'scope': 'profile email',
    'width': 240,
    'height': 50,
    'theme': 'dark',
    'onsuccess': onSignIn,
    'onfailure': onFailure
  });
}
