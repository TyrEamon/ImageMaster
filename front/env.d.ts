/// <reference types="vite/client" />

declare module "opencc-js/core" {
  export type Converter = (input: string) => string;

  export function ConverterFactory(from: unknown, to: unknown): Converter;
}

declare module "opencc-js/preset" {
  export const from: Record<string, unknown>;
  export const to: Record<string, unknown>;
}
