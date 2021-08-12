$(() => {

  function initTags() {
    const tagsInputs = $(".tags-input")

    tagsInputs.select2({
      tags: true,
      width: '300px',
      minimumInputLength: 1,
      tokenSeparators: [','],
      matcher: (params, data) => {
        return null
      }
    }).on('select2:select', e => {
      const tag = e.params.data.id
      if (tag == null) {
        return
      }

      const url = $(e.target).attr('data-url')
      $.ajax({
        url: url,
        type: 'POST',
        data: JSON.stringify({ tag: tag }),
        dataType: 'json',
        success: (data) => {
          let o = new Option(tag, tag);
          $(o).html(tag);
          $("#tags_filter").append(o);
          $("#tags_filter").selectpicker('refresh')
        }
      })

    }).on('select2:unselect', e => {
      const url = $(e.target).attr('data-url') + "/" + e.params.data.id
      $.ajax({
        url: url,
        type: 'DELETE'
      })
    }).on('select2:open', function (e) {
      $('.select2-container--open .select2-dropdown--below').css('display', 'none')
    })
  }

  initTags()

  $(window).on('table:reloaded', e => {
    initTags()
  })
})
