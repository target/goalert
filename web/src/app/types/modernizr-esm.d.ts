// modernizr-esm currently does not include types.
// Instead, we leverage @types/modernizr, and expose the types
// on a per-module basis.
declare module 'modernizr-esm/feature/inputtypes' {
  import * as m from 'modernizr'
  export const inputtypes = m.inputtypes
}
