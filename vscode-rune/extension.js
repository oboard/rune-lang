const vscode = require("vscode");
const cp = require("child_process");
const fs = require("fs");
const path = require("path");
const { LanguageClient } = require("vscode-languageclient/node");

let client;
let outputChannel;
let runTerminal;
const revealOutputChannelOnError = 3;

function activate(context) {
  outputChannel = vscode.window.createOutputChannel("Rune Language Server");
  context.subscriptions.push(outputChannel);
  context.subscriptions.push(
    vscode.languages.registerCodeLensProvider(
      { scheme: "file", language: "rune" },
      new RuneCodeLensProvider()
    ),
    vscode.debug.registerDebugAdapterDescriptorFactory(
      "rune",
      new RuneDebugAdapterDescriptorFactory()
    ),
    vscode.debug.registerDebugConfigurationProvider(
      "rune",
      new RuneDebugConfigurationProvider()
    ),
    vscode.commands.registerCommand("rune.runFile", runFile),
    vscode.commands.registerCommand("rune.debugFile", debugFile),
    vscode.commands.registerCommand("rune.restartLanguageServer", async () => {
      await stopClient();
      startClient(context);
    }),
    vscode.commands.registerCommand("rune.showLanguageServerOutput", () => {
      outputChannel.show(true);
    })
  );
  startClient(context);
}

function deactivate() {
  return stopClient();
}

module.exports = { activate, deactivate };

class RuneCodeLensProvider {
  provideCodeLenses(document) {
    const lenses = [];
    for (let line = 0; line < document.lineCount; line++) {
      const text = document.lineAt(line).text;
      if (!/^\s*main\s*\(/.test(text)) {
        continue;
      }
      const range = new vscode.Range(line, 0, line, text.length);
      lenses.push(
        new vscode.CodeLens(range, {
          title: "$(play) Run",
          command: "rune.runFile",
          arguments: [document.uri]
        }),
        new vscode.CodeLens(range, {
          title: "$(debug-alt) Debug",
          command: "rune.debugFile",
          arguments: [document.uri]
        })
      );
    }
    return lenses;
  }
}

async function runFile(uri) {
  const target = await resolveTargetFile(uri);
  if (!target) {
    return;
  }
  const terminal = runeTerminal(target);
  terminal.show(true);
  terminal.sendText(`${shellQuote(target.command)} run ${shellQuote(target.file)}`);
}

async function debugFile(uri) {
  const target = await resolveTargetFile(uri);
  if (!target) {
    return;
  }
  await vscode.debug.startDebugging(vscode.workspace.getWorkspaceFolder(target.uri), {
    type: "rune",
    request: "launch",
    name: `Debug ${path.basename(target.file)}`,
    program: target.file,
    runeCommand: target.command,
    runeRoot: target.runeRoot,
    cwd: target.cwd
  });
}

async function resolveTargetFile(uri) {
  const documentUri = resolveDocumentUri(uri);
  if (!documentUri) {
    vscode.window.showWarningMessage("Open a Rune file first.");
    return undefined;
  }
  if (documentUri.scheme !== "file") {
    vscode.window.showWarningMessage("Rune can only run local .rn files.");
    return undefined;
  }
  const doc = await vscode.workspace.openTextDocument(documentUri);
  if (doc.isDirty) {
    await doc.save();
  }
  const runeRoot = resolveRuneRootForPath(documentUri.fsPath);
  return {
    uri: documentUri,
    file: documentUri.fsPath,
    runeRoot,
    command: resolveRuneCommand(runeRoot),
    cwd: runeRoot || workspaceRoot() || path.dirname(documentUri.fsPath)
  };
}

function resolveDocumentUri(uri) {
  if (uri instanceof vscode.Uri) {
    return uri;
  }
  return vscode.window.activeTextEditor?.document.uri;
}

function runeTerminal(target) {
  if (!runTerminal || runTerminal.exitStatus) {
    runTerminal = vscode.window.createTerminal({
      name: "Rune",
      cwd: target.cwd,
      env: {
        ...(target.runeRoot ? { RUNE_ROOT: target.runeRoot } : {})
      }
    });
  }
  return runTerminal;
}

class RuneDebugConfigurationProvider {
  resolveDebugConfiguration(_folder, config) {
    if (!config.type && !config.request && !config.name) {
      config.type = "rune";
      config.request = "launch";
      config.name = "Debug Rune File";
    }
    config.type = config.type || "rune";
    config.request = config.request || "launch";
    config.name = config.name || "Debug Rune File";
    if (!config.program) {
      const active = vscode.window.activeTextEditor?.document.uri;
      if (active?.scheme === "file") {
        config.program = active.fsPath;
      }
    }
    if (!config.program) {
      vscode.window.showWarningMessage("Open a Rune file before debugging.");
      return undefined;
    }
    const runeRoot = resolveRuneRootForPath(config.program);
    config.runeRoot = config.runeRoot || runeRoot;
    config.runeCommand = config.runeCommand || resolveRuneCommand(config.runeRoot);
    config.cwd = config.cwd || config.runeRoot || workspaceRoot() || path.dirname(config.program);
    return config;
  }
}

class RuneDebugAdapterDescriptorFactory {
  createDebugAdapterDescriptor() {
    return new vscode.DebugAdapterInlineImplementation(new RuneDebugAdapter());
  }
}

class RuneDebugAdapter {
  constructor() {
    this._onDidSendMessage = new vscode.EventEmitter();
    this.onDidSendMessage = this._onDidSendMessage.event;
    this.seq = 1;
    this.child = undefined;
  }

  handleMessage(message) {
    switch (message.command) {
      case "initialize":
        this.sendResponse(message, {
          supportsConfigurationDoneRequest: true,
          supportsTerminateRequest: true
        });
        this.sendEvent("initialized");
        break;
      case "launch":
        this.launch(message);
        break;
      case "setBreakpoints":
        this.sendResponse(message, {
          breakpoints: (message.arguments?.breakpoints || []).map((bp) => ({
            verified: false,
            line: bp.line,
            message: "Rune breakpoints are not wired to generated Go yet."
          }))
        });
        break;
      case "configurationDone":
      case "threads":
        if (message.command === "threads") {
          this.sendResponse(message, { threads: [{ id: 1, name: "Rune" }] });
        } else {
          this.sendResponse(message);
        }
        break;
      case "stackTrace":
        this.sendResponse(message, { stackFrames: [], totalFrames: 0 });
        break;
      case "scopes":
        this.sendResponse(message, { scopes: [] });
        break;
      case "variables":
        this.sendResponse(message, { variables: [] });
        break;
      case "disconnect":
      case "terminate":
        this.killChild();
        this.sendResponse(message);
        this.sendEvent("terminated");
        break;
      default:
        this.sendResponse(message);
        break;
    }
  }

  launch(request) {
    const args = request.arguments || {};
    const program = args.program;
    if (!program) {
      this.sendResponse(request, undefined, false, "Missing Rune program.");
      return;
    }
    const command = args.runeCommand || "rune";
    const runeArgs = ["run", program, ...normalizeArgs(args.args)];
    this.sendOutput(`$ ${[command, ...runeArgs].map(shellQuote).join(" ")}\n`, "console");
    this.child = cp.spawn(command, runeArgs, {
      cwd: args.cwd || path.dirname(program),
      env: {
        ...process.env,
        ...(args.runeRoot ? { RUNE_ROOT: args.runeRoot } : {})
      }
    });
    this.sendResponse(request);
    if (this.child.pid) {
      this.sendEvent("process", {
        name: path.basename(program),
        systemProcessId: this.child.pid,
        isLocalProcess: true,
        startMethod: "launch"
      });
    }
    this.child.stdout.on("data", (data) => this.sendOutput(data.toString(), "stdout"));
    this.child.stderr.on("data", (data) => this.sendOutput(data.toString(), "stderr"));
    this.child.on("error", (error) => {
      this.sendOutput(`${error.message}\n`, "stderr");
      this.sendEvent("terminated");
    });
    this.child.on("exit", (code, signal) => {
      this.sendEvent("exited", { exitCode: code ?? 0 });
      if (signal) {
        this.sendOutput(`terminated by ${signal}\n`, "console");
      }
      this.sendEvent("terminated");
    });
  }

  dispose() {
    this.killChild();
    this._onDidSendMessage.dispose();
  }

  killChild() {
    if (this.child && !this.child.killed) {
      this.child.kill();
    }
    this.child = undefined;
  }

  sendResponse(request, body, success = true, message = undefined) {
    const response = {
      seq: this.seq++,
      type: "response",
      request_seq: request.seq,
      success,
      command: request.command
    };
    if (body !== undefined) {
      response.body = body;
    }
    if (message) {
      response.message = message;
    }
    this._onDidSendMessage.fire(response);
  }

  sendEvent(event, body = undefined) {
    const message = {
      seq: this.seq++,
      type: "event",
      event
    };
    if (body !== undefined) {
      message.body = body;
    }
    this._onDidSendMessage.fire(message);
  }

  sendOutput(output, category) {
    this.sendEvent("output", { category, output });
  }
}

function startClient(context) {
  const runeRoot = resolveRuneRoot();
  const command = resolveRuneCommand(runeRoot);
  const serverOptions = {
    command,
    args: ["lsp"],
    options: {
      cwd: runeRoot || workspaceRoot(),
      env: {
        ...process.env,
        ...(runeRoot ? { RUNE_ROOT: runeRoot } : {})
      }
    }
  };
  const clientOptions = {
    documentSelector: [{ scheme: "file", language: "rune" }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/*.rn")
    },
    outputChannel,
    revealOutputChannelOn: revealOutputChannelOnError
  };
  outputChannel.appendLine(`Starting Rune LSP: ${command} lsp`);
  if (runeRoot) {
    outputChannel.appendLine(`Rune root: ${runeRoot}`);
  }
  client = new LanguageClient("rune", "Rune Language Server", serverOptions, clientOptions);
  const disposable = client.start();
  context.subscriptions.push(disposable);
}

async function stopClient() {
  if (!client) {
    return undefined;
  }
  const running = client;
  client = undefined;
  return running.stop();
}

function resolveRuneCommand(runeRoot) {
  const configured = vscode.workspace.getConfiguration("rune").get("serverPath");
  if (configured && configured.trim() !== "") {
    return configured;
  }

  const root = runeRoot || resolveRuneRoot();
  if (root) {
    const candidate = path.join(root, ".bin", executableName("rune"));
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  return "rune";
}

function resolveRuneRootForPath(filePath) {
  const configured = vscode.workspace.getConfiguration("rune").get("root");
  if (configured && configured.trim() !== "") {
    return configured;
  }
  if (filePath) {
    const root = findRuneRoot(path.dirname(filePath));
    if (root) {
      return root;
    }
  }
  return resolveRuneRoot();
}

function resolveRuneRoot() {
  const configured = vscode.workspace.getConfiguration("rune").get("root");
  if (configured && configured.trim() !== "") {
    return configured;
  }

  for (const folder of vscode.workspace.workspaceFolders || []) {
    const root = findRuneRoot(folder.uri.fsPath);
    if (root) {
      return root;
    }
  }

  const active = vscode.window.activeTextEditor?.document.uri;
  if (active?.scheme === "file") {
    return findRuneRoot(path.dirname(active.fsPath));
  }

  return undefined;
}

function findRuneRoot(start) {
  let dir = start;
  while (dir && dir !== path.dirname(dir)) {
    if (
      fs.existsSync(path.join(dir, "go.mod")) &&
      fs.existsSync(path.join(dir, "core")) &&
      fs.existsSync(path.join(dir, "cmd", "rune"))
    ) {
      return dir;
    }
    dir = path.dirname(dir);
  }
  return undefined;
}

function workspaceRoot() {
  return vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
}

function executableName(name) {
  return process.platform === "win32" ? `${name}.exe` : name;
}

function normalizeArgs(args) {
  if (!args) {
    return [];
  }
  return Array.isArray(args) ? args.map(String) : [String(args)];
}

function shellQuote(value) {
  const text = String(value);
  if (process.platform === "win32") {
    return `"${text.replace(/"/g, '\\"')}"`;
  }
  return `'${text.replace(/'/g, `'\\''`)}'`;
}
