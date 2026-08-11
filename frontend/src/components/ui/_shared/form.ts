// Form contract (types + injection key) — the machinery lives in form/useForm.ts.
// This file is always installed so PrtFormField can check for form context without
// depending on the form component (Shape A as a dependency-graph fact).
import type { InjectionKey } from 'vue'

// Standard Schema v1 — inlined on purpose; the spec recommends inlining over depending on @standard-schema/spec.
export interface StandardSchemaV1<Input = unknown, Output = Input> {
  readonly '~standard': StandardSchemaV1.Props<Input, Output>
}

export declare namespace StandardSchemaV1 {
  export interface Props<Input = unknown, Output = Input> {
    readonly version: 1
    readonly vendor: string
    // Takes `unknown`; may return sync OR async — always await it.
    readonly validate: (
      value: unknown,
    ) => Result<Output> | Promise<Result<Output>>
    readonly types?: Types<Input, Output> | undefined
  }
  export type Result<Output> = SuccessResult<Output> | FailureResult
  export interface SuccessResult<Output> {
    readonly value: Output
    readonly issues?: undefined
  }
  export interface FailureResult {
    readonly issues: ReadonlyArray<Issue>
  }
  export interface Issue {
    readonly message: string
    // Segments are PropertyKey OR { key } — Valibot uses the object form.
    readonly path?: ReadonlyArray<PropertyKey | PathSegment> | undefined
  }
  export interface PathSegment {
    readonly key: PropertyKey
  }
  export interface Types<Input = unknown, Output = Input> {
    readonly input: Input
    readonly output: Output
  }
  export type InferInput<S extends StandardSchemaV1> = NonNullable<
    S['~standard']['types']
  >['input']
  export type InferOutput<S extends StandardSchemaV1> = NonNullable<
    S['~standard']['types']
  >['output']
}

// 'user.email', 'tags.0' — normalizes BOTH path-segment forms.
export function toDotPath(
  path: ReadonlyArray<PropertyKey | StandardSchemaV1.PathSegment>,
): string {
  return path
    .map((seg) =>
      String(
        typeof seg === 'object' && seg !== null && 'key' in seg ? seg.key : seg,
      ),
    )
    .join('.')
}

// The form context — what PrtFormField sees via inject().
export interface PrtFormError {
  // dot-path: 'title', 'user.email', 'tags.0'
  name: string
  message: string
}

export interface PrtFormContext {
  // dot-path → first message. Raw — NOT display-gated. Reactive.
  readonly errors: Readonly<Record<string, string>>
  readonly submitCount: number
  readonly validating: boolean
  readonly submitting: boolean
  // Display-gated error for a field (reward-early-punish-late applied).
  displayError(name: string, pattern?: RegExp): string | undefined
  // Field protocol — PrtFormField wires these via inputProps listeners.
  reportBlur(name: string): void
  reportInput(name: string): void
  validate(): Promise<boolean>
  submit(): Promise<void>
  setErrors(
    errors: PrtFormError[] | Record<string, string>,
    mode?: 'merge' | 'replace',
  ): void
  clearErrors(name?: string): void
  // Clears errors + blurred set + submitCount. State is the user's; untouched.
  reset(): void
}

export const prtFormKey = Symbol('prt-form') as InjectionKey<PrtFormContext>
