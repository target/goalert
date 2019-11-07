export default function copyToClipboard(text) {
  if (!navigator.clipboard) {
    window.alert(
      'Copying to clipboard is not supported in this browser. Please use Chrome or Firefox.',
    )
    return
  }
  navigator.clipboard.writeText(text)
}
