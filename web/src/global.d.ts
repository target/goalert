/* eslint-disable camelcase */
declare namespace NodeJS {
  export interface Global {
    __webpack_public_path__: string
    pathPrefix: string
    GOALERT_VERSION: string
  }

  declare module '*.md'
}

// ElementType yields a union of values present in a const array
// e.g. const xyz = [1, "2", 3.5] as const
// type XYZ = ElementType<typeof counts> === "2" | 1 | 3.5
export type ElementType<
  T extends ReadonlyArray<unknown>
> = T extends ReadonlyArray<infer ElementType> ? ElementType : never
