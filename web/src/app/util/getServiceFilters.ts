export default function getServiceFilters(input: string): {
  labelKey: string
  labelValue: string
  integrationKey: string
} {
  // grab key and value from the input param, if at all
  let labelKey = ''
  let labelValue = ''
  let integrationKey = ''
  if (input.includes('token=')) {
    const tokenStr = input.substring(0, 42)
    integrationKey = tokenStr.slice(6)
    input = input.replace(tokenStr, '').trim() // remove token string from input
  }
  if (input.includes('=')) {
    const searchSplit = input.split(/(!=|=)/)
    labelKey = searchSplit[0]
    // the value can contain "=", so joining the rest of the match such that it doesn't get lost
    labelValue = searchSplit.slice(2).join('')
  }
  console.log('here,', labelKey, labelValue)

  return { labelKey, labelValue, integrationKey }
}
