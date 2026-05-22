const vscode = require("vscode");
const fs = require("fs");
const path = require("path");
const { LanguageClient } = require("vscode-languageclient/node");

let client;
let outputChannel;
const revealOutputChannelOnError = 3;

function activate(context) {
  outputChannel = vscode.window.createOutputChannel("Rune Language Server");
  context.subscriptions.push(outputChannel);
  context.subscriptions.push(
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
