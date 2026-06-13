import Editor, { loader } from "@monaco-editor/react";
import {
  AlertTriangle,
  Braces,
  CheckCircle2,
  Eraser,
  FileCode2,
  Sparkles,
  Terminal,
} from "lucide-react";
import * as esbuild from "esbuild-wasm";
import esbuildWasmURL from "esbuild-wasm/esbuild.wasm?url";
import * as monaco from "monaco-editor/esm/vs/editor/editor.api.js";
import "monaco-editor/esm/vs/editor/contrib/codelens/browser/codelensController.js";
import "monaco-editor/esm/vs/editor/contrib/documentSymbols/browser/documentSymbols.js";
import "monaco-editor/esm/vs/editor/contrib/format/browser/formatActions.js";
import "monaco-editor/esm/vs/editor/contrib/gotoSymbol/browser/goToCommands.js";
import "monaco-editor/esm/vs/editor/contrib/hover/browser/hoverContribution.js";
import "monaco-editor/esm/vs/editor/contrib/inlayHints/browser/inlayHintsContribution.js";
import "monaco-editor/esm/vs/editor/contrib/rename/browser/rename.js";
import "monaco-editor/esm/vs/editor/contrib/semanticTokens/browser/documentSemanticTokens.js";
import "monaco-editor/esm/vs/editor/contrib/suggest/browser/suggestController.js";
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";

type MonacoModule = typeof monaco;
type MonacoEditor = monaco.editor.IStandaloneCodeEditor;
type Disposable = { dispose(): void };
type MonacoEnvironmentConfig = {
  getWorker(moduleId: string, label: string): Worker;
};

type GoRuntime = {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
};

declare global {
  interface Window {
    Go?: new () => GoRuntime;
    runeCompile?: (source: string) => string;
    runeFormat?: (source: string) => string;
    runeInitCompiler?: (coreSources: string) => string;
    runeLSP?: (request: string) => string;
    runeTest?: (request: string) => string;
    __runeLanguageDisposables?: Disposable[];
  }
}

type Example = {
  name: string;
  label: string;
  source: string;
  runnable: boolean;
  testable: boolean;
};

type RuneDiagnostic = {
  message: string;
  line: number;
  column: number;
};

type BridgeResponse = {
  ok: boolean;
  error?: string;
  diagnostics?: RuneDiagnostic[];
  lsp?: unknown;
  tests?: RuneTestResult[];
  typescript?: string;
  entries?: RuntimeEntries;
  formatted?: string;
  elapsedMs?: number;
};

type RuntimeEntries = {
  main?: string;
  render?: string;
};

type RuneTestResult = {
  name: string;
  passed: boolean;
  error?: string;
  output?: string;
  elapsedMs?: number;
};

type CompileState = {
  phase: "loading" | "ready" | "compiling" | "error";
  message: string;
  diagnostics: RuneDiagnostic[];
  elapsedMs?: number;
};

type ConsoleEntry = {
  id: string;
  level: "log" | "warn" | "error" | "system";
  text: string;
};

type OutputPanel = "typescript" | "console" | "preview";
type RuntimeEntryMode = "main" | "render" | "auto";

type RuntimeMessage = {
  token: string;
  kind: "console" | "done";
  level?: ConsoleEntry["level"];
  values?: string[];
};

type LSPMethod =
  | "diagnostics"
  | "hover"
  | "completion"
  | "definition"
  | "references"
  | "codeLens"
  | "documentSymbol"
  | "formatting"
  | "inlayHint"
  | "semanticTokens"
  | "rename";

type LSPPosition = {
  line: number;
  character: number;
};

type LSPRange = {
  start: LSPPosition;
  end: LSPPosition;
};

type LSPCommand = {
  title: string;
  command: string;
  arguments?: unknown[];
};

type LSPCodeLens = {
  range: LSPRange;
  command?: LSPCommand;
};

type RuneTestSpec = {
  uri?: string;
  name: string;
  line?: number;
  character?: number;
};

type LSPTextEdit = {
  range: LSPRange;
  newText: string;
};

type LSPLocation = {
  uri: string;
  range: LSPRange;
};

type LSPHover = {
  contents: {
    value: string;
  };
  range?: LSPRange;
};

type LSPCompletionItem = {
  label: string;
  detail?: string;
  kind?: number;
};

type LSPDocumentSymbol = {
  name: string;
  detail?: string;
  kind: number;
  range: LSPRange;
  selectionRange: LSPRange;
};

type LSPInlayHint = {
  position: LSPPosition;
  label: string;
  kind?: number;
  tooltip?: string;
};

type LSPSemanticTokens = {
  data?: number[];
};

type LSPWorkspaceEdit = {
  changes?: Record<string, LSPTextEdit[]>;
};

type LSPRequestOptions = {
  position?: LSPPosition;
  includeDeclaration?: boolean;
  newName?: string;
  tabSize?: number;
  insertSpaces?: boolean;
};

const monacoGlobal = globalThis as typeof globalThis & {
  MonacoEnvironment?: MonacoEnvironmentConfig;
};

monacoGlobal.MonacoEnvironment = {
  getWorker(_moduleId: string, _label: string) {
    return new editorWorker();
  },
};

loader.config({ monaco });

const exampleModules = import.meta.glob<string>("../../examples/*.rn", {
  eager: true,
  import: "default",
  query: "?raw",
});

const coreSourceModules = import.meta.glob<string>("../../core/*/*.rn", {
  eager: true,
  import: "default",
  query: "?raw",
});

const exampleOrder = [
  "fib",
  "array",
  "map",
  "json",
  "signal",
  "counter",
  "enum",
  "struct",
  "regex",
  "async_wait",
  "anonymous_object",
  "complex_type",
  "complex_type2",
  "nested_match",
  "dyn_list",
  "list",
  "user",
  "freedom",
  "leetcode_top100",
];

const hiddenExampleNames = new Set(["ffi"]);
const runnableExampleNames = new Set(exampleOrder.filter((name) => name !== "leetcode_top100"));

const examples = Object.entries(exampleModules)
  .map(([path, source]) => {
    const name = path.split("/").pop()?.replace(/\.rn$/, "") ?? path;
    return {
      name,
      label: toTitle(name),
      source,
      runnable: runnableExampleNames.has(name),
      testable: sourceHasTests(source),
    };
  })
  .filter((example) => !hiddenExampleNames.has(example.name))
  .sort((left, right) => {
    const leftIndex = exampleOrder.indexOf(left.name);
    const rightIndex = exampleOrder.indexOf(right.name);
    if (leftIndex >= 0 && rightIndex >= 0) {
      return leftIndex - rightIndex;
    }
    if (leftIndex >= 0) {
      return -1;
    }
    if (rightIndex >= 0) {
      return 1;
    }
    return left.label.localeCompare(right.label);
  });

const defaultExample = examples.find((example) => example.name === "fib") ?? examples[0];
const fallbackSource = 'main() => @io.println("Hello, Rune")';
const initialExample = exampleFromLocation() ?? defaultExample;
const initialSource = initialExample?.source ?? fallbackSource;
const initialExampleName = initialExample?.name ?? "";

const coreSources = Object.fromEntries(Object.entries(coreSourceModules));

let runeWasmPromise: Promise<void> | undefined;
let esbuildPromise: Promise<void> | undefined;
let runeLanguageConfigured = false;
let runMainCommandRegistered = false;
let runMainCommandHandler: (() => void) | undefined;
let previewRenderCommandRegistered = false;
let previewRenderCommandHandler: (() => void) | undefined;
let runTestCommandRegistered = false;
let runTestCommandHandler: ((spec?: unknown) => void) | undefined;
let entryCounter = 0;

function App() {
  const [source, setSource] = useState(initialSource);
  const [selectedExample, setSelectedExample] = useState(initialExampleName);
  const [typescriptOutput, setTypeScriptOutput] = useState("");
  const [compileState, setCompileState] = useState<CompileState>({
    phase: "loading",
    message: "Loading Rune compiler",
    diagnostics: [],
  });
  const [activePanel, setActivePanel] = useState<OutputPanel>("console");
  const [consoleEntries, setConsoleEntries] = useState<ConsoleEntry[]>([
    {
      id: nextEntryId(),
      level: "system",
      text: "Ready to compile Rune in the browser.",
    },
  ]);
  const [isRunning, setIsRunning] = useState(false);
  const editorRef = useRef<MonacoEditor | null>(null);
  const monacoRef = useRef<MonacoModule | null>(null);
  const iframeRef = useRef<HTMLIFrameElement | null>(null);
  const compileSequence = useRef(0);
  const runCurrentRef = useRef<() => void>(() => {});
  const previewCurrentRef = useRef<() => void>(() => {});
  const runTestCurrentRef = useRef<(spec?: unknown) => void>(() => {});
  const previewSourceRef = useRef("");

  const currentExample = useMemo(
    () => examples.find((example) => example.name === selectedExample),
    [selectedExample],
  );

  const diagnosticsCount = compileState.diagnostics.length;
  const statusTone = compileState.phase === "error" || diagnosticsCount > 0 ? "bad" : "good";

  const applyDiagnostics = useCallback((diagnostics: RuneDiagnostic[]) => {
    const editor = editorRef.current;
    const monacoModule = monacoRef.current;
    const model = editor?.getModel();
    if (!model || !monacoModule) {
      return;
    }
    monacoModule.editor.setModelMarkers(
      model,
      "rune",
      diagnostics.map((diag) => {
        const startLineNumber = Math.max(1, diag.line || 1);
        const startColumn = Math.max(1, diag.column || 1);
        return {
          severity: monacoModule.MarkerSeverity.Error,
          message: diag.message,
          startLineNumber,
          startColumn,
          endLineNumber: startLineNumber,
          endColumn: startColumn + 1,
        };
      }),
    );
  }, []);

  const compileSource = useCallback(
    async (nextSource: string, silent = false) => {
      const sequence = compileSequence.current + 1;
      compileSequence.current = sequence;
      if (!silent) {
        setCompileState((state) => ({
          ...state,
          phase: "compiling",
          message: "Compiling Rune",
        }));
      }
      try {
        const result = await compileRune(nextSource);
        if (sequence !== compileSequence.current) {
          return result;
        }
        const diagnostics = result.diagnostics ?? [];
        setTypeScriptOutput(result.typescript ?? "");
        applyDiagnostics(diagnostics);
        setCompileState({
          phase: result.ok ? "ready" : "error",
          message: result.ok ? "Compiled to TypeScript" : result.error ?? "Compile failed",
          diagnostics,
          elapsedMs: result.elapsedMs,
        });
        return result;
      } catch (error) {
        const message = errorMessage(error);
        if (sequence === compileSequence.current) {
          setCompileState({
            phase: "error",
            message,
            diagnostics: [],
          });
          applyDiagnostics([]);
        }
        return {
          ok: false,
          error: message,
          diagnostics: [],
        } satisfies BridgeResponse;
      }
    },
    [applyDiagnostics],
  );

  useEffect(() => {
    const handle = window.setTimeout(() => {
      void compileSource(source, true);
    }, 320);
    return () => window.clearTimeout(handle);
  }, [compileSource, source]);

  useEffect(() => {
    void ensureRuneWasm().then(
      () => {
        setCompileState((state) =>
          state.phase === "loading"
            ? {
                ...state,
                phase: "ready",
                message: "Rune compiler ready",
              }
            : state,
        );
      },
      (error: unknown) => {
        setCompileState({
          phase: "error",
          message: errorMessage(error),
          diagnostics: [],
        });
      },
    );
  }, []);

  useEffect(() => {
    const routeExample = exampleFromLocation();
    if (!routeExample && initialExampleName) {
      replaceExampleRoute(initialExampleName);
    }

    const handlePopState = () => {
      const nextExample = exampleFromLocation() ?? defaultExample;
      if (!nextExample) {
        return;
      }
      setSelectedExample(nextExample.name);
      setSource(nextExample.source);
      setActivePanel("console");
      previewSourceRef.current = "";
      setConsoleEntries([
        {
          id: nextEntryId(),
          level: "system",
          text: `Loaded ${nextExample.label}.`,
        },
      ]);
      clearPreview(iframeRef.current);
    };

    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const handleBeforeMount = useCallback((monacoModule: MonacoModule) => {
    configureRuneLanguage(monacoModule);
  }, []);

  const handleMount = useCallback(
    (editor: MonacoEditor, monacoModule: MonacoModule) => {
      editorRef.current = editor;
      monacoRef.current = monacoModule;
      runMainCommandHandler = () => {
        runCurrentRef.current();
      };
      if (!runMainCommandRegistered) {
        runMainCommandRegistered = true;
        monacoModule.editor.registerCommand("rune.runMain", () => {
          runMainCommandHandler?.();
        });
      }
      previewRenderCommandHandler = () => {
        previewCurrentRef.current();
      };
      if (!previewRenderCommandRegistered) {
        previewRenderCommandRegistered = true;
        monacoModule.editor.registerCommand("rune.previewRender", () => {
          previewRenderCommandHandler?.();
        });
      }
      runTestCommandHandler = (spec?: unknown) => {
        runTestCurrentRef.current(spec);
      };
      if (!runTestCommandRegistered) {
        runTestCommandRegistered = true;
        monacoModule.editor.registerCommand("rune.runTest", (_accessor, spec?: unknown) => {
          runTestCommandHandler?.(spec);
        });
      }
      editor.addAction({
        id: "rune.runMain",
        label: "Rune: Run Main",
        run: () => {
          runCurrentRef.current();
        },
      });
      editor.addAction({
        id: "rune.previewRender",
        label: "Rune: Preview Render",
        run: () => {
          previewCurrentRef.current();
        },
      });
      applyDiagnostics(compileState.diagnostics);
    },
    [applyDiagnostics, compileState.diagnostics],
  );

  const selectExample = useCallback((example: Example) => {
    pushExampleRoute(example.name);
    setSelectedExample(example.name);
    setSource(example.source);
    setActivePanel("console");
    previewSourceRef.current = "";
    setConsoleEntries([
      {
        id: nextEntryId(),
        level: "system",
        text: `Loaded ${example.label}.`,
      },
    ]);
    clearPreview(iframeRef.current);
  }, []);

  const formatCurrent = useCallback(async () => {
    const result = await formatRune(source);
    if (result.ok && result.formatted !== undefined) {
      setSource(result.formatted);
      previewSourceRef.current = "";
      clearPreview(iframeRef.current);
      setConsoleEntries((entries) => [
        ...entries,
        {
          id: nextEntryId(),
          level: "system",
          text: `Formatted in ${formatDuration(result.elapsedMs)}.`,
        },
      ]);
      return;
    }
    const diagnostics = result.diagnostics ?? [];
    applyDiagnostics(diagnostics);
    setCompileState({
      phase: "error",
      message: result.error ?? "Format failed",
      diagnostics,
      elapsedMs: result.elapsedMs,
    });
    setActivePanel("console");
    setConsoleEntries((entries) => [
      ...entries,
      {
        id: nextEntryId(),
        level: "error",
        text: diagnostics.length > 0 ? diagnostics[0].message : result.error ?? "Format failed",
      },
    ]);
  }, [applyDiagnostics, source]);

  const executeCurrent = useCallback(
    async (entryMode: RuntimeEntryMode, panel: "console" | "preview") => {
      if (isRunning) {
        return;
      }
      setIsRunning(true);
      setActivePanel(panel);
      setConsoleEntries([
        {
          id: nextEntryId(),
          level: "system",
          text: "Compiling Rune source.",
        },
      ]);
      try {
        const result = await compileSource(source);
        if (!result.ok || !result.typescript) {
          setConsoleEntries((entries) => [
            ...entries,
            {
              id: nextEntryId(),
              level: "error",
              text: result.error ?? "Compile failed",
            },
          ]);
          return;
        }
        setConsoleEntries((entries) => [
          ...entries,
          {
            id: nextEntryId(),
            level: "system",
            text: entryMode === "render" ? "Rendering preview." : "Running generated JavaScript.",
          },
        ]);
        await executeTypeScript(result.typescript, result.entries, iframeRef.current, entryMode, (entry) => {
          setConsoleEntries((entries) => [...entries, entry]);
        });
        if (entryMode === "render") {
          previewSourceRef.current = source;
        }
      } catch (error) {
        setActivePanel("console");
        setConsoleEntries((entries) => [
          ...entries,
          {
            id: nextEntryId(),
            level: "error",
            text: errorMessage(error),
          },
        ]);
      } finally {
        setIsRunning(false);
      }
    },
    [compileSource, isRunning, source],
  );

  const runCurrent = useCallback(async () => {
    await executeCurrent("main", "console");
  }, [executeCurrent]);

  const previewCurrent = useCallback(async () => {
    await executeCurrent("render", "preview");
  }, [executeCurrent]);

  const runTestCurrent = useCallback(
    async (spec?: unknown) => {
      if (isRunning) {
        return;
      }
      const testSpec = isRuneTestSpec(spec) ? spec : undefined;
      setIsRunning(true);
      setActivePanel("console");
      setConsoleEntries([
        {
          id: nextEntryId(),
          level: "system",
          text: testSpec?.name ? `Running test "${testSpec.name}".` : "Running tests.",
        },
      ]);
      try {
        const result = await runRuneTest(source, testSpec?.name, testSpec?.uri);
        applyDiagnostics(result.diagnostics ?? []);
        const tests = result.tests ?? [];
        const entries = testResultsToConsoleEntries(tests);
        if (!result.ok && tests.length === 0) {
          entries.push({
            id: nextEntryId(),
            level: "error",
            text: result.error ?? "Test failed.",
          });
        }
        entries.push({
          id: nextEntryId(),
          level: result.ok ? "system" : "error",
          text: testSummaryText(tests, result.elapsedMs),
        });
        setConsoleEntries((current) => [...current, ...entries]);
      } catch (error) {
        setConsoleEntries((current) => [
          ...current,
          {
            id: nextEntryId(),
            level: "error",
            text: errorMessage(error),
          },
        ]);
      } finally {
        setIsRunning(false);
      }
    },
    [applyDiagnostics, isRunning, source],
  );

  useEffect(() => {
    runCurrentRef.current = () => {
      void runCurrent();
    };
  }, [runCurrent]);

  useEffect(() => {
    previewCurrentRef.current = () => {
      void previewCurrent();
    };
  }, [previewCurrent]);

  useEffect(() => {
    runTestCurrentRef.current = (spec?: unknown) => {
      void runTestCurrent(spec);
    };
  }, [runTestCurrent]);

  const handleSourceChange = useCallback((value?: string) => {
    setSource(value ?? "");
    previewSourceRef.current = "";
    clearPreview(iframeRef.current);
  }, []);

  const openPreview = useCallback(() => {
    setActivePanel("preview");
    if (sourceHasRender(source) && previewSourceRef.current !== source) {
      previewCurrentRef.current();
    }
  }, [source]);

  const clearConsole = useCallback(() => {
    setConsoleEntries([]);
  }, []);

  return (
    <main className="playground-shell">
      <section className="workspace">
        <aside className="examples-pane" aria-label="Examples">
          <div className="pane-heading">
            <FileCode2 size={16} />
            <span>Examples</span>
          </div>
          <div className="mobile-example-picker">
            <select
              aria-label="Select example"
              className="mobile-example-select"
              value={selectedExample}
              onChange={(event) => {
                const example = examples.find((item) => item.name === event.target.value);
                if (example) {
                  selectExample(example);
                }
              }}
            >
              {examples.map((example) => (
                <option key={example.name} value={example.name}>
                  {example.label}
                  {example.runnable ? "" : example.testable ? " (test)" : " (check)"}
                </option>
              ))}
            </select>
          </div>
          <div className="examples-list">
            {examples.map((example) => (
              <button
                type="button"
                key={example.name}
                className={example.name === selectedExample ? "example-row active" : "example-row"}
                onClick={() => selectExample(example)}
              >
                <span>{example.label}</span>
                <small>{example.runnable ? "TS" : example.testable ? "test" : "check"}</small>
              </button>
            ))}
          </div>
        </aside>

        <section className="editor-pane" aria-label="Rune source editor">
          <div className="pane-toolbar">
            <div>
              <span className="toolbar-title">{currentExample?.label ?? "Scratch"}</span>
              <span className={`status-chip ${statusTone}`}>
                {statusTone === "good" ? <CheckCircle2 size={14} /> : <AlertTriangle size={14} />}
                {compileState.message}
              </span>
            </div>
            <div className="editor-toolbar-actions">
              <span className="diagnostic-count">
                {diagnosticsCount === 0 ? "0 diagnostics" : `${diagnosticsCount} diagnostics`}
                {compileState.elapsedMs !== undefined ? ` · ${formatDuration(compileState.elapsedMs)}` : ""}
              </span>
              <button type="button" className="ghost-button compact-button" onClick={() => void formatCurrent()}>
                <Sparkles size={15} />
                Format
              </button>
            </div>
          </div>
          <div className="editor-frame">
            <Editor
              beforeMount={handleBeforeMount}
              defaultLanguage="rune"
              language="rune"
              onChange={handleSourceChange}
              onMount={handleMount}
              options={{
                automaticLayout: true,
                codeLens: true,
                fontFamily: "JetBrains Mono, SFMono-Regular, Consolas, monospace",
                fontSize: 14,
                formatOnPaste: true,
                formatOnType: false,
                glyphMargin: false,
                lineNumbersMinChars: 3,
                minimap: { enabled: false },
                padding: { bottom: 16, top: 16 },
                renderLineHighlight: "all",
                scrollBeyondLastLine: false,
                tabSize: 2,
                wordWrap: "on",
              }}
              path="playground.rn"
              theme="rune-dark"
              value={source}
            />
          </div>
        </section>

        <aside className="output-pane" aria-label="Output">
          <div className="output-tabs" role="tablist" aria-label="Output panels">
            <button
              type="button"
              className={activePanel === "typescript" ? "active" : ""}
              onClick={() => setActivePanel("typescript")}
            >
              <Braces size={15} />
              TypeScript
            </button>
            <button
              type="button"
              className={activePanel === "console" ? "active" : ""}
              onClick={() => setActivePanel("console")}
            >
              <Terminal size={15} />
              Console
            </button>
            <button
              type="button"
              className={activePanel === "preview" ? "active" : ""}
              onClick={openPreview}
            >
              <FileCode2 size={15} />
              Preview
            </button>
          </div>

          <div className={activePanel === "typescript" ? "panel-body active" : "panel-body"}>
            <pre className="typescript-output">
              {typescriptOutput || "Compile a Rune example to view generated TypeScript."}
            </pre>
          </div>

          <div className={activePanel === "console" ? "panel-body active" : "panel-body"}>
            <div className="console-toolbar">
              <span>{consoleEntries.length} entries</span>
              <button type="button" onClick={clearConsole}>
                <Eraser size={14} />
                Clear
              </button>
            </div>
            <div className="console-output">
              {consoleEntries.length === 0 ? (
                <p className="empty-output">No console output.</p>
              ) : (
                consoleEntries.map((entry) => (
                  <div key={entry.id} className={`console-line ${entry.level}`}>
                    <span>{entry.level}</span>
                    <code>{entry.text}</code>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className={activePanel === "preview" ? "panel-body active preview-body" : "panel-body preview-body"}>
            <iframe
              ref={iframeRef}
              className="preview-frame"
              sandbox="allow-scripts"
              title="Rune runtime preview"
            />
          </div>
        </aside>
      </section>
    </main>
  );
}

async function ensureRuneWasm() {
  if (runeWasmPromise) {
    return runeWasmPromise;
  }
  runeWasmPromise = loadRuneWasm().catch((error: unknown) => {
    runeWasmPromise = undefined;
    throw error;
  });
  return runeWasmPromise;
}

async function loadRuneWasm() {
  await loadScript(playgroundAssetPath("wasm_exec.js"));
  if (!window.Go) {
    throw new Error("Go wasm runtime did not register window.Go");
  }
  const go = new window.Go();
  const wasmResponse = await fetch(playgroundAssetPath("rune.wasm"));
  if (!wasmResponse.ok) {
    throw new Error(`Failed to load rune.wasm: ${wasmResponse.status}`);
  }
  const wasmBuffer = await wasmResponse.arrayBuffer();
  const wasm = await WebAssembly.instantiate(wasmBuffer, go.importObject);
  void go.run(wasm.instance);
  if (!window.runeInitCompiler) {
    throw new Error("Rune wasm bridge did not register runeInitCompiler");
  }
  const init = parseBridgeResponse(window.runeInitCompiler(JSON.stringify(coreSources)));
  if (!init.ok) {
    throw new Error(init.error ?? "Rune compiler initialization failed");
  }
}

function playgroundAssetPath(name: string) {
  const base = import.meta.env.BASE_URL || "/";
  return `${base.endsWith("/") ? base : `${base}/`}${name}`;
}

async function compileRune(source: string) {
  await ensureRuneWasm();
  if (!window.runeCompile) {
    throw new Error("Rune compiler bridge is unavailable");
  }
  return parseBridgeResponse(window.runeCompile(source));
}

async function formatRune(source: string) {
  await ensureRuneWasm();
  if (!window.runeFormat) {
    throw new Error("Rune formatter bridge is unavailable");
  }
  return parseBridgeResponse(window.runeFormat(source));
}

async function runRuneTest(source: string, name?: string, uri = "file:///playground.rn") {
  await ensureRuneWasm();
  if (!window.runeTest) {
    throw new Error("Rune test bridge is unavailable");
  }
  return parseBridgeResponse(
    window.runeTest(
      JSON.stringify({
        uri,
        source,
        name: name ?? "",
      }),
    ),
  );
}

async function requestRuneLSP<T>(
  source: string,
  uri: string,
  method: LSPMethod,
  options: LSPRequestOptions = {},
) {
  await ensureRuneWasm();
  if (!window.runeLSP) {
    throw new Error("Rune LSP bridge is unavailable");
  }
  const result = parseBridgeResponse(
    window.runeLSP(
      JSON.stringify({
        method,
        uri,
        source,
        line: options.position?.line ?? 0,
        character: options.position?.character ?? 0,
        includeDeclaration: options.includeDeclaration ?? false,
        newName: options.newName ?? "",
        tabSize: options.tabSize ?? 2,
        insertSpaces: options.insertSpaces ?? true,
      }),
    ),
  );
  if (!result.ok) {
    throw new Error(result.error ?? `Rune LSP ${method} failed`);
  }
  return result.lsp as T | undefined;
}

async function ensureEsbuild() {
  if (esbuildPromise) {
    return esbuildPromise;
  }
  esbuildPromise = esbuild
    .initialize({
      wasmURL: esbuildWasmURL,
      worker: true,
    })
    .catch((error: unknown) => {
      esbuildPromise = undefined;
      throw error;
    });
  return esbuildPromise;
}

async function executeTypeScript(
  typescript: string,
  entries: RuntimeEntries | undefined,
  frame: HTMLIFrameElement | null,
  entryMode: RuntimeEntryMode,
  appendConsoleEntry: (entry: ConsoleEntry) => void,
) {
  if (!frame) {
    throw new Error("Runtime preview frame is unavailable");
  }
  await ensureEsbuild();
  const previewMountId = "rune-preview-root";
  const transformed = await esbuild.transform(`${typescript}\n\n${runtimeFooter(previewMountId, entryMode, entries)}`, {
    format: "esm",
    loader: "ts",
    sourcemap: "inline",
    target: "es2022",
  });
  const token = crypto.randomUUID();
  const messageHandler = (event: MessageEvent<unknown>) => {
    if (event.source !== frame.contentWindow || !isRuntimeMessage(event.data, token)) {
      return;
    }
    if (event.data.kind === "console") {
      appendConsoleEntry({
        id: nextEntryId(),
        level: event.data.level ?? "log",
        text: event.data.values?.join(" ") ?? "",
      });
    }
    if (event.data.kind === "done") {
      appendConsoleEntry({
        id: nextEntryId(),
        level: "system",
        text: "Runtime finished.",
      });
      window.removeEventListener("message", messageHandler);
    }
  };
  window.addEventListener("message", messageHandler);
  frame.srcdoc = runtimeDocument(transformed.code, token, previewMountId);
}

function runtimeFooter(previewMountId: string, entryMode: RuntimeEntryMode, entries?: RuntimeEntries) {
  const mainRef = runtimeEntryReference(entries?.main);
  const renderRef = runtimeEntryReference(entries?.render);
  return `
async function __runeRunMain() {
  const __runeMain = ${mainRef};
  if (typeof __runeMain !== "function") {
    return false;
  }
  const __runeMainResult = __runeMain();
  if (__runeMainResult && typeof __runeMainResult.then === "function") {
    await __runeMainResult;
  }
  return true;
}

async function __runeRunRender(__runePreview) {
  const __runeRender = ${renderRef};
  if (typeof __runeRender !== "function" || !__runePreview) {
    return false;
  }
  const __runeRenderResult = __runeRender();
  const __runeRendered =
    __runeRenderResult && typeof __runeRenderResult.then === "function"
      ? await __runeRenderResult
      : __runeRenderResult;
  __runePreview.replaceChildren(
    __runeRendered instanceof Node ? __runeRendered : document.createTextNode(String(__runeRendered)),
  );
  return true;
}

async function __runeStart() {
  const __runePreview = document.getElementById(${JSON.stringify(previewMountId)});
  const __runeEntryMode = ${JSON.stringify(entryMode)};
  if (__runeEntryMode === "main") {
    if (!(await __runeRunMain())) {
      console.log("No main() function found.");
    }
  } else if (__runeEntryMode === "render") {
    if (!(await __runeRunRender(__runePreview))) {
      console.log("No render() function found.");
    }
  } else if (!(await __runeRunMain()) && !(await __runeRunRender(__runePreview))) {
    console.log("No main() or render() function found.");
  }
  if (typeof runeWaitAll === "function") {
    await runeWaitAll();
  }
}

try {
  await __runeStart();
} catch (error) {
  console.error(error);
}
`;
}

function runtimeEntryReference(name?: string) {
  return typeof name === "string" && /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(name) ? name : "undefined";
}

function runtimeDocument(code: string, token: string, previewMountId: string) {
  const escapedCode = code.replaceAll("</script", "<\\/script");
  return `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <style>
      :root {
        color: #17211f;
        background: #f7faf7;
        font: 14px/1.45 ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      }
      body {
        margin: 0;
        min-height: 100vh;
        background:
          linear-gradient(135deg, rgba(37, 168, 125, 0.1), transparent 42%),
          #f7faf7;
      }
      #${previewMountId} {
        min-height: 100vh;
        padding: 20px;
        box-sizing: border-box;
      }
      button {
        border: 1px solid #9ccfbf;
        border-radius: 8px;
        background: #ffffff;
        color: #14352c;
        cursor: pointer;
        font: inherit;
        padding: 7px 12px;
      }
      button:hover {
        background: #e7f6ef;
      }
    </style>
  </head>
  <body>
    <div id="${previewMountId}"></div>
    <script type="module">
      const __runeToken = ${JSON.stringify(token)};
      const __runePost = (payload) => parent.postMessage({ token: __runeToken, ...payload }, "*");
      const __runeFormat = (value) => {
        if (value instanceof Error) return value.stack || value.message;
        if (typeof value === "string") return value;
        if (typeof value === "undefined") return "undefined";
        try {
          const json = JSON.stringify(value);
          return typeof json === "string" ? json : String(value);
        } catch {
          return String(value);
        }
      };
      const console = {
        log: (...values) => __runePost({ kind: "console", level: "log", values: values.map(__runeFormat) }),
        warn: (...values) => __runePost({ kind: "console", level: "warn", values: values.map(__runeFormat) }),
        error: (...values) => __runePost({ kind: "console", level: "error", values: values.map(__runeFormat) }),
      };
      window.addEventListener("error", (event) => {
        console.error(event.error || event.message);
        __runePost({ kind: "done" });
      });
      window.addEventListener("unhandledrejection", (event) => {
        console.error(event.reason);
        __runePost({ kind: "done" });
      });
${escapedCode}
      __runePost({ kind: "done" });
    </script>
  </body>
</html>`;
}

function clearPreview(frame: HTMLIFrameElement | null) {
  if (frame) {
    frame.srcdoc = "";
  }
}

function loadScript(src: string) {
  const existing = document.querySelector<HTMLScriptElement>(`script[src="${src}"]`);
  if (existing) {
    return Promise.resolve();
  }
  return new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.src = src;
    script.async = true;
    script.addEventListener("load", () => resolve(), { once: true });
    script.addEventListener("error", () => reject(new Error(`Failed to load ${src}`)), { once: true });
    document.head.append(script);
  });
}

function parseBridgeResponse(raw: string): BridgeResponse {
  try {
    return JSON.parse(raw) as BridgeResponse;
  } catch (error) {
    return {
      ok: false,
      error: errorMessage(error),
      diagnostics: [],
    };
  }
}

function configureRuneLanguage(monacoModule: MonacoModule) {
  if (runeLanguageConfigured) {
    return;
  }
  runeLanguageConfigured = true;
  window.__runeLanguageDisposables?.forEach((disposable) => disposable.dispose());
  const disposables: Disposable[] = [];
  window.__runeLanguageDisposables = disposables;
  const track = (disposable: Disposable | undefined | void) => {
    if (disposable) {
      disposables.push(disposable);
    }
  };
  track(monacoModule.languages.register({ id: "rune" }));
  track(monacoModule.languages.setLanguageConfiguration("rune", {
    brackets: [
      ["{", "}"],
      ["[", "]"],
      ["(", ")"],
    ],
    comments: {
      lineComment: "//",
    },
    autoClosingPairs: [
      { open: "{", close: "}" },
      { open: "[", close: "]" },
      { open: "(", close: ")" },
      { open: '"', close: '"' },
    ],
  }));
  track(monacoModule.languages.setMonarchTokensProvider("rune", {
    keywords: [
      "async",
      "def",
      "enum",
      "false",
      "func",
      "go",
      "match",
      "null",
      "return",
      "test",
      "this",
      "true",
    ],
    operators: ["=>", "->", ":=", "$=", "?", "??", "==", "!=", "<=", ">=", "+", "-", "*", "/", "%"],
    tokenizer: {
      root: [
        [/[A-Z][\w$]*/, "type.identifier"],
        [/[a-z_$][\w$]*/, { cases: { "@keywords": "keyword", "@default": "identifier" } }],
        [/@[a-zA-Z_][\w$]*/, "annotation"],
        [/\/\/.*$/, "comment"],
        [/"([^"\\]|\\.)*$/, "string.invalid"],
        [/"/, "string", "@string"],
        [/\d+\.\d+/, "number.float"],
        [/\d+/, "number"],
        [/[{}()[\]]/, "@brackets"],
        [/[+\-*/%!=<>?:|&]+/, "operator"],
      ],
      string: [
        [/[^\\"]+/, "string"],
        [/\\./, "string.escape"],
        [/"/, "string", "@pop"],
      ],
    },
  }));
  monacoModule.editor.defineTheme("rune-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "7dd3c7" },
      { token: "annotation", foreground: "f0b85c" },
      { token: "type.identifier", foreground: "9bbcf7" },
      { token: "string", foreground: "b8e986" },
      { token: "number", foreground: "f7c67c" },
      { token: "comment", foreground: "6f7f7a" },
    ],
    colors: {
      "editor.background": "#111715",
      "editor.foreground": "#dce7e1",
      "editor.lineHighlightBackground": "#1a2421",
      "editorLineNumber.foreground": "#596862",
      "editorLineNumber.activeForeground": "#b6d7ca",
      "editorCursor.foreground": "#72e0bd",
      "editor.selectionBackground": "#245d4d",
      "editor.inactiveSelectionBackground": "#1e3933",
    },
  });
  track(monacoModule.languages.registerCompletionItemProvider("rune", {
    triggerCharacters: ["@", "."],
    async provideCompletionItems(model, position) {
      const word = model.getWordUntilPosition(position);
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      };
      const items = await requestRuneLSP<LSPCompletionItem[]>(
        model.getValue(),
        model.uri.toString(),
        "completion",
        { position: toLSPPosition(position) },
      ).catch(() => []);
      return {
        suggestions: normalizeCompletionItems(items).map((item) => ({
          label: item.label,
          kind: completionKind(monacoModule, item.kind),
          insertText: item.label,
          detail: item.detail,
          range,
        })),
      };
    },
  }));
  track(monacoModule.languages.registerHoverProvider("rune", {
    async provideHover(model, position) {
      const hover = await requestRuneLSP<LSPHover>(model.getValue(), model.uri.toString(), "hover", {
        position: toLSPPosition(position),
      }).catch(() => undefined);
      if (!hover?.contents?.value) {
        return null;
      }
      return {
        contents: [{ value: hover.contents.value }],
        range: hover.range ? toMonacoRange(monacoModule, hover.range) : undefined,
      };
    },
  }));
  track(monacoModule.languages.registerDefinitionProvider("rune", {
    async provideDefinition(model, position) {
      const location = await requestRuneLSP<LSPLocation>(model.getValue(), model.uri.toString(), "definition", {
        position: toLSPPosition(position),
      }).catch(() => undefined);
      return location ? toMonacoLocation(monacoModule, location) : null;
    },
  }));
  track(monacoModule.languages.registerReferenceProvider("rune", {
    async provideReferences(model, position) {
      const locations =
        (await requestRuneLSP<LSPLocation[]>(model.getValue(), model.uri.toString(), "references", {
          position: toLSPPosition(position),
          includeDeclaration: true,
        }).catch(() => [])) ?? [];
      return locations.map((location) => toMonacoLocation(monacoModule, location));
    },
  }));
  track(monacoModule.languages.registerDocumentSymbolProvider("rune", {
    async provideDocumentSymbols(model) {
      const symbols =
        (await requestRuneLSP<LSPDocumentSymbol[]>(
          model.getValue(),
          model.uri.toString(),
          "documentSymbol",
        ).catch(() => [])) ?? [];
      return symbols.map((symbol) => ({
        name: symbol.name,
        detail: symbol.detail ?? "",
        kind: symbolKind(monacoModule, symbol.kind),
        tags: [],
        range: toMonacoRange(monacoModule, symbol.range),
        selectionRange: toMonacoRange(monacoModule, symbol.selectionRange),
      }));
    },
  }));
  track(monacoModule.languages.registerDocumentFormattingEditProvider("rune", {
    async provideDocumentFormattingEdits(model) {
      const edits =
        (await requestRuneLSP<LSPTextEdit[]>(model.getValue(), model.uri.toString(), "formatting", {
          tabSize: 2,
          insertSpaces: true,
        }).catch(() => [])) ?? [];
      return edits.map((edit) => ({
        range: toMonacoRange(monacoModule, edit.range),
        text: edit.newText,
      }));
    },
  }));
  track(monacoModule.languages.registerInlayHintsProvider("rune", {
    async provideInlayHints(model) {
      const hints =
        (await requestRuneLSP<LSPInlayHint[]>(model.getValue(), model.uri.toString(), "inlayHint").catch(
          () => [],
        )) ?? [];
      return {
        hints: hints.map((hint) => ({
          position: toMonacoPosition(monacoModule, hint.position),
          label: hint.label,
          kind: monacoModule.languages.InlayHintKind.Type,
          tooltip: hint.tooltip,
          paddingLeft: true,
        })),
        dispose() {},
      };
    },
  }));
  track(monacoModule.languages.registerDocumentSemanticTokensProvider("rune", {
    getLegend() {
      return {
        tokenTypes: ["variable", "type", "function"],
        tokenModifiers: ["modification", "async"],
      };
    },
    async provideDocumentSemanticTokens(model) {
      const tokens =
        (await requestRuneLSP<LSPSemanticTokens>(
        model.getValue(),
        model.uri.toString(),
        "semanticTokens",
        ).catch(() => ({ data: [] }))) ?? { data: [] };
      return {
        data: Uint32Array.from(tokens.data ?? []),
      };
    },
    releaseDocumentSemanticTokens() {},
  }));
  track(monacoModule.languages.registerRenameProvider("rune", {
    async provideRenameEdits(model, position, newName) {
      const edit = await requestRuneLSP<LSPWorkspaceEdit>(model.getValue(), model.uri.toString(), "rename", {
        position: toLSPPosition(position),
        newName,
      }).catch(() => undefined);
      return toMonacoWorkspaceEdit(monacoModule, edit);
    },
  }));
  track(monacoModule.languages.registerCodeLensProvider("rune", {
    async provideCodeLenses(model) {
      const lspLenses =
        (await requestRuneLSP<LSPCodeLens[]>(model.getValue(), model.uri.toString(), "codeLens").catch(
          () => [],
        )) ?? [];
      const lenses: monaco.languages.CodeLens[] = [];
      for (const lspLens of lspLenses) {
        const lens = toMonacoCodeLens(monacoModule, lspLens);
        if (lens) {
          lenses.push(lens);
        }
      }
      lenses.push(
        ...mainRunCodeLenses(monacoModule, model).map((range) => ({
          range,
          command: {
            id: "rune.runMain",
            title: "$(play) Run",
            arguments: [],
          },
        })),
        ...renderPreviewCodeLenses(monacoModule, model).map((range) => ({
          range,
          command: {
            id: "rune.previewRender",
            title: "$(preview) Preview",
            arguments: [],
          },
        })),
      );
      return {
        lenses,
        dispose() {},
      };
    },
  }));
}

function toMonacoCodeLens(monacoModule: MonacoModule, lens: LSPCodeLens): monaco.languages.CodeLens | undefined {
  if (!lens.command) {
    return undefined;
  }
  return {
    range: toMonacoRange(monacoModule, lens.range),
    command: {
      id: lens.command.command,
      title: lens.command.title,
      arguments: lens.command.arguments,
    },
  };
}

function mainRunCodeLenses(monacoModule: MonacoModule, model: monaco.editor.ITextModel) {
  const ranges: monaco.Range[] = [];
  for (let lineNumber = 1; lineNumber <= model.getLineCount(); lineNumber += 1) {
    const line = model.getLineContent(lineNumber);
    if (!/^\s*main\s*\(/.test(line)) {
      continue;
    }
    ranges.push(new monacoModule.Range(lineNumber, 1, lineNumber, line.length + 1));
  }
  return ranges;
}

function renderPreviewCodeLenses(monacoModule: MonacoModule, model: monaco.editor.ITextModel) {
  const ranges: monaco.Range[] = [];
  for (let lineNumber = 1; lineNumber <= model.getLineCount(); lineNumber += 1) {
    const line = model.getLineContent(lineNumber);
    if (!/^\s*render\s*\(/.test(line)) {
      continue;
    }
    ranges.push(new monacoModule.Range(lineNumber, 1, lineNumber, line.length + 1));
  }
  return ranges;
}

function sourceHasRender(value: string) {
  return /^\s*render\s*\(/m.test(value);
}

function sourceHasTests(value: string) {
  return /^\s*\?\s*"/m.test(value);
}

function isRuneTestSpec(value: unknown): value is RuneTestSpec {
  if (!isRecord(value) || typeof value.name !== "string") {
    return false;
  }
  return value.name.length > 0;
}

function testResultsToConsoleEntries(results: RuneTestResult[]) {
  return results.flatMap((result) => {
    const entries: ConsoleEntry[] = [
      {
        id: nextEntryId(),
        level: result.passed ? "system" : "error",
        text: `${result.passed ? "PASS" : "FAIL"} ${result.name} (${formatDuration(result.elapsedMs)})`,
      },
    ];
    if (result.output) {
      entries.push({
        id: nextEntryId(),
        level: "log",
        text: result.output.trimEnd(),
      });
    }
    if (result.error) {
      entries.push({
        id: nextEntryId(),
        level: "error",
        text: result.error,
      });
    }
    return entries;
  });
}

function testSummaryText(results: RuneTestResult[], elapsedMs?: number) {
  if (results.length === 0) {
    return "No tests ran.";
  }
  const passed = results.filter((result) => result.passed).length;
  const failed = results.length - passed;
  const status = failed === 0 ? "PASS" : "FAIL";
  return `${status} ${results.length} test${results.length === 1 ? "" : "s"}, ${passed} passed, ${failed} failed in ${formatDuration(elapsedMs)}.`;
}

function toLSPPosition(position: monaco.Position): LSPPosition {
  return {
    line: Math.max(0, position.lineNumber - 1),
    character: Math.max(0, position.column - 1),
  };
}

function toMonacoPosition(monacoModule: MonacoModule, position: LSPPosition) {
  return new monacoModule.Position(position.line + 1, position.character + 1);
}

function toMonacoRange(monacoModule: MonacoModule, range: LSPRange) {
  return new monacoModule.Range(
    range.start.line + 1,
    range.start.character + 1,
    range.end.line + 1,
    range.end.character + 1,
  );
}

function toMonacoLocation(monacoModule: MonacoModule, location: LSPLocation) {
  return {
    uri: monacoModule.Uri.parse(location.uri),
    range: toMonacoRange(monacoModule, location.range),
  };
}

function toMonacoWorkspaceEdit(monacoModule: MonacoModule, edit?: LSPWorkspaceEdit) {
  const edits = Object.entries(edit?.changes ?? {}).flatMap(([uri, textEdits]) =>
    textEdits.map((textEdit) => ({
      resource: monacoModule.Uri.parse(uri),
      textEdit: {
        range: toMonacoRange(monacoModule, textEdit.range),
        text: textEdit.newText,
      },
      versionId: undefined,
    })),
  );
  return { edits };
}

function normalizeCompletionItems(items?: LSPCompletionItem[]) {
  return (items ?? []).filter((item) => item.label.length > 0);
}

function completionKind(monacoModule: MonacoModule, kind?: number) {
  switch (kind) {
    case 6:
      return monacoModule.languages.CompletionItemKind.Variable;
    case 7:
      return monacoModule.languages.CompletionItemKind.Class;
    case 14:
      return monacoModule.languages.CompletionItemKind.Keyword;
    default:
      return monacoModule.languages.CompletionItemKind.Function;
  }
}

function symbolKind(monacoModule: MonacoModule, kind: number) {
  switch (kind) {
    case 5:
      return monacoModule.languages.SymbolKind.Class;
    case 13:
      return monacoModule.languages.SymbolKind.Variable;
    case 23:
      return monacoModule.languages.SymbolKind.Struct;
    default:
      return monacoModule.languages.SymbolKind.Function;
  }
}

function isRuntimeMessage(value: unknown, token: string): value is RuntimeMessage {
  if (!isRecord(value)) {
    return false;
  }
  return value.token === token && (value.kind === "console" || value.kind === "done");
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function errorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

function formatDuration(value?: number) {
  if (value === undefined) {
    return "";
  }
  if (value < 1) {
    return `${Math.round(value * 1000)}us`;
  }
  return `${value.toFixed(1)}ms`;
}

function nextEntryId() {
  entryCounter += 1;
  return `entry-${entryCounter}`;
}

function exampleFromLocation() {
  const name = exampleNameFromLocation();
  if (!name) {
    return undefined;
  }
  return examples.find((example) => example.name === name);
}

function exampleNameFromLocation() {
  const relativePath = relativeRoutePath();
  if (!relativePath) {
    return undefined;
  }
  const parts = relativePath.split("/").filter(Boolean).map(decodePathPart);
  if (parts[0] === "examples" && parts[1]) {
    return parts[1];
  }
  if (parts.length === 1 && examples.some((example) => example.name === parts[0])) {
    return parts[0];
  }
  return undefined;
}

function pushExampleRoute(exampleName: string) {
  const route = exampleRoute(exampleName);
  if (window.location.pathname === route) {
    return;
  }
  window.history.pushState({ example: exampleName }, "", route);
}

function replaceExampleRoute(exampleName: string) {
  window.history.replaceState({ example: exampleName }, "", exampleRoute(exampleName));
}

function exampleRoute(exampleName: string) {
  return `${baseRoutePath()}examples/${encodeURIComponent(exampleName)}`;
}

function relativeRoutePath() {
  const base = baseRoutePath();
  const pathname = window.location.pathname;
  if (base !== "/" && pathname.startsWith(base)) {
    return pathname.slice(base.length).replace(/^\/+/, "");
  }
  return pathname.replace(/^\/+/, "");
}

function baseRoutePath() {
  const base = import.meta.env.BASE_URL || "/";
  if (base === "/") {
    return "/";
  }
  return `/${base.replace(/^\/+|\/+$/g, "")}/`;
}

function decodePathPart(part: string) {
  try {
    return decodeURIComponent(part);
  } catch {
    return part;
  }
}

function toTitle(name: string) {
  return name
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export default App;
