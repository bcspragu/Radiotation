$(function() {
  $('form').submit(function() {
    var form = $(this);
    $.ajax({
      url: form.attr('action'),
      method: form.attr('method'),
      data: form.serialize(),
      success: function(resp) {
        if (resp) {
          alert(resp);
        }
      }
    });

    return false;
  });
});
