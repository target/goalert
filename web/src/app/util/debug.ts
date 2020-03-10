export function debug(args: any) {
  if (!window.console) return // some browsers don't define console if the devtools are closed
  if (console.debug) console.debug(args)
  // not all browsers have `.debug` defined
  else console.log(args)
}
