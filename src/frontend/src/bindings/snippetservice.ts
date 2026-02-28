// Manual v3 bindings for SnippetService.
// These will be replaced by auto-generated bindings from `wails3 generate bindings`.

import { Call } from "@wailsio/runtime";

const pkg = "github.com/infktd/snipt/src/internal/gui";
const svc = "SnippetService";

function call(method: string, ...args: any[]): Promise<any> {
  return Call.ByName(`${pkg}.${svc}.${method}`, ...args);
}

export function ListSnippets(opts: any) {
  return call("ListSnippets", opts);
}

export function SearchSnippets(query: string) {
  return call("SearchSnippets", query);
}

export function CreateSnippet(snippet: any) {
  return call("CreateSnippet", snippet);
}

export function UpdateSnippet(snippet: any) {
  return call("UpdateSnippet", snippet);
}

export function UpdateSnippetTags(id: string, tags: string[]) {
  return call("UpdateSnippetTags", id, tags);
}

export function DeleteSnippet(id: string) {
  return call("DeleteSnippet", id);
}

export function SetPinned(id: string, pinned: boolean) {
  return call("SetPinned", id, pinned);
}

export function IncrementUseCount(id: string) {
  return call("IncrementUseCount", id);
}

export function GetStats() {
  return call("GetStats");
}

export function GetConfig() {
  return call("GetConfig");
}

export function UpdateConfig(cfg: any) {
  return call("UpdateConfig", cfg);
}

export function GetDBPath(): Promise<string> {
  return call("GetDBPath");
}

export function GetVersion(): Promise<string> {
  return call("GetVersion");
}

export function SyncSetup(token: string) {
  return call("SyncSetup", token);
}

export function SyncNow() {
  return call("SyncNow");
}

export function SyncStatus() {
  return call("SyncStatus");
}

export function SyncDisconnect() {
  return call("SyncDisconnect");
}

export function IsSyncConfigured(): Promise<boolean> {
  return call("IsSyncConfigured");
}

export function GetLaunchAtLogin(): Promise<boolean> {
  return call("GetLaunchAtLogin");
}

export function SetLaunchAtLogin(enabled: boolean): Promise<void> {
  return call("SetLaunchAtLogin", enabled);
}
