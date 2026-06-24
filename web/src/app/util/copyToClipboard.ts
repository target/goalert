import { isIOS } from './browsers'

/*
 * Creates a temporary text area element, selects the
 * text inside, and copies it to the clipboard.
 *
 * If other text was highlighted before this operation,
 * that state is saved upon completion of copying.
 */
function fallback(str: string): void {
  const textArea = document.createElement('textArea') as HTMLTextAreaElement
  textArea.value = str // Set its value to what you want copied
  textArea.readOnly = true // Deny tampering
  document.body.appendChild(textArea)

  // Check if there is any content selected previously
  const docSelection = document.getSelection()
  const selected =
    docSelection && docSelection.rangeCount > 0
      ? docSelection.getRangeAt(0) // Store selection if found
      : false

  // iOS requires some special finesse
  if (isIOS) {
    const range = document.createRange()
    range.selectNodeContents(textArea)
    const windowSelection = window.getSelection()

    if (windowSelection) {
      windowSelection.removeAllRanges()
      windowSelection.addRange(range)
    } else {
      return console.error('Failed to copy')
    }

    textArea.setSelectionRange(0, 999999) // Big number to copy everything
  } else {
    textArea.select()
  }

  document.execCommand('copy') // Execute copy command as a result of some event
  document.body.removeChild(textArea) // Remove text area from the HTML document

  // If a selection existed before copying
  if (selected) {
    let docSelection = document.getSelection()
    if (docSelection) docSelection.removeAllRanges() // Unselect everything on the HTML document

    docSelection = document.getSelection()
    if (docSelection) docSelection.addRange(selected) // Restore the original selection
  }
}

export default function copyToClipboard(text: string): void {
  try {
    navigator.clipboard.writeText(text)
  } catch {
    fallback(text)
  }
}
