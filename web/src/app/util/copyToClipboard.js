/*
 * Creates a temporary text area element, selects the
 * text inside, and copies it to the clipboard.
 *
 * If other text was highlighted before this operation,
 * that state is saved upon completion of copying.
 */
export default function copyToClipboard(str) {
  const textArea = document.createElement('textArea')
  textArea.value = str // Set its value to what you want copied
  textArea.readOnly = true // Deny tampering
  document.body.appendChild(textArea)

  // Check if there is any content selected previously
  const selected =
    document.getSelection().rangeCount > 0
      ? document.getSelection().getRangeAt(0) // Store selection if found
      : false

  // iOS requires some special finesse
  if (isOS()) {
    let range = document.createRange()
    range.selectNodeContents(textArea)
    let selection = window.getSelection()
    selection.removeAllRanges()
    selection.addRange(range)
    textArea.setSelectionRange(0, 999999) // Big number to copy everything
  } else {
    textArea.select()
  }

  document.execCommand('copy') // Execute copy command as a result of some event
  document.body.removeChild(textArea) // Remove text area from the HTML document

  // If a selection existed before copying
  if (selected) {
    document.getSelection().removeAllRanges() // Unselect everything on the HTML document
    document.getSelection().addRange(selected) // Restore the original selection
  }
}

function isOS() {
  return navigator.userAgent.match(/ipad|ipod|iphone/i)
}
