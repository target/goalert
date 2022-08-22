function liveUpdate(...selector) {
  setInterval(function () {
    fetch(window.location.href)
      .then((value) => value.text())
      .then((data) => {
        var template = document.createElement('template')
        template.innerHTML = data.trim()

        selector.forEach((s) => {
          const docEl = document.querySelector(s)
          const tmplEl = template.content.querySelector(s)

          if (docEl.innerHTML === tmplEl.innerHTML) return

          docEl.replaceWith(tmplEl)
        })
      })
  }, 1000)
}
