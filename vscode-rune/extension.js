const vscode = require("vscode");
const fs = require("fs");
const path = require("path");
const { LanguageClient, TransportKind } = require("vscode-languageclient/node");

let client;

function activate(context) {
  const command = resolveRuneCommand();
  const serverOptions = {
    command,
    args: ["lsp"],
    transport: TransportKind.stdio
  };
  const clientOptions = {
    documentSelector: [{ scheme: "file", language: "rune" }]
  };
  client = new LanguageClient("rune", "Rune Language Server", serverOptions, clientOptions);
  context.subscriptions.push(client.start());
}

function deactivate() {
  if (!client) {
    return undefined;
  }
  return client.stop();
}

module.exports = { activate, deactivate };

function resolveRuneCommand() {
  const configured = vscode.workspace.getConfiguration("rune").get("serverPath");
  if (configured && configured.trim() !== "") {
    return configured;
  }

  for (const folder of vscode.workspace.workspaceFolders || []) {
    const candidate = path.join(folder.uri.fsPath, ".bin", process.platform === "win32" ? "rune.exe" : "rune");
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  return "rune";
}
