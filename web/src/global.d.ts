/* eslint-disable camelcase */
declare namespace NodeJS {
  export interface Global {
    __webpack_public_path__: string
    pathPrefix: string
    GOALERT_VERSION: string
  }

  declare module '*.md'
}

export type ElementType<
  T extends ReadonlyArray<unknown>
> = T extends ReadonlyArray<infer ElementType> ? ElementType : never
