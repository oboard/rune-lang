import type { Plugin } from "vite";

export type RuneMatcher = string | RegExp | ((id: string) => boolean) | Array<string | RegExp | ((id: string) => boolean)>;

export interface RunePluginOptions {
  include?: RuneMatcher;
  exclude?: RuneMatcher;
  declaration?: boolean;
  runtime?: "browser" | "node";
  runeBin?: string | string[];
  runeRoot?: string;
}

export declare function rune(options?: RunePluginOptions): Plugin;
export default rune;
