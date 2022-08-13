function liveUpdate(selector) {
  setInterval(function () {
    fetch(window.location.href)
      .then((value) => value.text())
      .then((data) => {
        var template = document.createElement('template')
        template.innerHTML = data.trim()
        document
          .querySelector(selector)
          .replaceWith(template.content.querySelector(selector))
      })
  }, 1000)
}
