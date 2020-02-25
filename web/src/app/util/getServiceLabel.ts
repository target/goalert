export default function getServiceLabel(input: string) {
  // grab key and value from the input param, if at all
  let labelKey = ''
  let labelValue = ''
  if (input.includes('=')) {
    const searchSplit = input.split(/(!=|=)/)
    labelKey = searchSplit[0]
    // the value can contain "=", so joining the rest of the match such that it doesn't get lost
    labelValue = searchSplit.slice(2).join('')
  }

  return { labelKey, labelValue }
}
