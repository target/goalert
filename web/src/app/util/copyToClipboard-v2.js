export default text => {
  if (!navigator.clipboard) {
    window.alert(
      'Copying to clipboard is not supported in this browser. Please use Chrome or Firefox.',
    )
    return
  }
  navigator.clipboard.writeText(text).then(
    function() {
      console.log('Async: Copying to clipboard was successful!')
    },
    function(err) {
      console.error('Async: Could not copy text: ', err)
    },
  )
}
