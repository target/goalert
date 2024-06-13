import type { CombinedError } from 'urql'

type ErrorConsumerError = {
  path: string
  code: string
  fieldID: string
  message: string
}
type ErrorStore = {
  errors: Set<ErrorConsumerError>
}

/** ErrorConsumer is a utility class for consuming and handling errors from a CombinedError. */
export class ErrorConsumer {
  constructor(e?: CombinedError | null | undefined) {
    if (!e) return

    if (e.networkError) {
      this.store.errors.add({
        message: e.networkError.message,
        code: '',
        path: '',
        fieldID: '',
      })
    }

    e.graphQLErrors?.forEach((e) => {
      const path = e.path?.join('.') || ''

      if (e.extensions.isFieldError) {
        this.store.errors.add({
          message: e.message,
          code: '_LEGACY_FIELD_ERROR',
          fieldID: e.extensions.fieldName?.toString() || '',
          path,
        })
        return
      }

      if (e.extensions.isMultiFieldError) {
        type fieldError = {
          fieldName: string
          message: string
        }
        const errs = (e.extensions.fieldErrors || []) as Array<fieldError>
        errs.forEach((fe: fieldError) => {
          this.store.errors.add({
            message: fe.message,
            code: '_LEGACY_FIELD_ERROR',
            fieldID: fe.fieldName,
            path,
          })
        })
        return
      }

      this.store.errors.add({
        message: e.message,
        code: e.extensions?.code?.toString() || '',
        path: e.path?.join('.') || '',
        fieldID: e.extensions?.fieldID?.toString() || '',
      })
    })

    this.hadErrors = this.hasErrors()

    // Use FinalizationRegistry if available, this will allow us to raise any errors that are forgotten.
    if (
      'FinalizationRegistry' in window &&
      typeof window.FinalizationRegistry === 'function'
    ) {
      // @ts-expect-error FinalizationRegistry is not in the lib
      const r = new window.FinalizationRegistry((e: { store: ErrorStore }) => {
        if (e.store.errors.size === 0) return
        e.store.errors.forEach((e) => console.error(e))
      })

      r.register(this, { e: this.store }, this)
    }
  }

  private isDone: boolean = false
  private store: ErrorStore = { errors: new Set() }

  /** Whether there were any errors in the original error. */
  public readonly hadErrors: boolean = false

  private doneCheck(): void {
    if (!this.isDone) return

    throw new Error(
      'ErrorConsumer is already done, ensure you are not calling this method after calling done() or remaining()',
    )
  }

  /** Returns and consumes (if exists) a single field error with the given name.
   *
   * Name maps to the `fname` when using the backend `validate.*` methods.
   */
  getFieldError(name: string): string | undefined {
    this.doneCheck()

    let result: string | undefined

    this.store.errors.forEach((e) => {
      if (e.code !== '_LEGACY_FIELD_ERROR') return
      if (e.fieldID !== name) return
      if (result !== undefined) return

      result = e.message
      this.store.errors.delete(e)
    })

    return result
  }

  /** Returns and consumes (if exists) a single INVALID_INPUT_VALUE error with the given path. */
  getInputError(path: string): string | undefined {
    this.doneCheck()

    let result: string | undefined

    // find the first error with the given path
    this.store.errors.forEach((e) => {
      if (e.code !== 'INVALID_INPUT_VALUE') return
      if (e.path !== path) return
      if (result !== undefined) return

      result = e.message
      this.store.errors.delete(e)
    })

    return result
  }

  /** Returns and consumes (if exists) all INVALID_DEST_FIELD_VALUE errors.
   *
   * @param path - If provided, only errors with the given path will be consumed.
   */
  getAllDestFieldErrors(path?: string): Readonly<Record<string, string>> {
    this.doneCheck()

    const errs: Record<string, string> = {}
    this.store.errors.forEach((e) => {
      if (e.code !== 'INVALID_DEST_FIELD_VALUE') return
      if (path !== undefined && e.path !== path) return
      if (errs[e.fieldID] !== undefined) return

      errs[e.fieldID] = e.message
      this.store.errors.delete(e)
    })

    return errs
  }

  /** Returns and consumes (if exists) all INVALID_DYNAMIC_PARAM_VALUE errors.
   *
   * @param path - If provided, only errors with the given path will be consumed.
   */
  getAllDynamicParamErrors(path?: string): Readonly<Record<string, string>> {
    this.doneCheck()

    const errs: Record<string, string> = {}
    this.store.errors.forEach((e) => {
      if (e.code !== 'INVALID_DYNAMIC_PARAM_VALUE') return
      if (pathPrefix !== undefined && !e.path.startsWith(pathPrefix)) return
      if (errs[e.fieldID] !== undefined) return

      errs[e.fieldID] = e.message
      this.store.errors.delete(e)
    })

    return errs
  }

  /** Returns whether there are any errors remaining. */
  hasErrors(): boolean {
    return this.store.errors.size > 0
  }

  /** Returns and consumes all remaining errors. */
  remaining(): Readonly<Array<string>> {
    this.doneCheck()

    const errs: string[] = []
    this.store.errors.forEach((e) => {
      errs.push(e.message)
      this.store.errors.delete(e)
    })

    // mark as done
    this.isDone = true

    return errs
  }

  /** Returns all remaining errors as an array of objects with a message key.
   *
   * Suitable for use with FormDialog.
   */
  remainingLegacy(): Array<{ message: string }> {
    return this.remaining().map((e) => ({ message: e }))
  }

  /** Logs and consumes any remaining errors. */
  done(): void {
    this.doneCheck()
    this.remaining().forEach((e) => console.error(e))
  }
}

/** useErrorConsumer is a hook for creating an ErrorConsumer. */
export function useErrorConsumer(
  e?: CombinedError | null | undefined,
): ErrorConsumer {
  return new ErrorConsumer(e)
}
