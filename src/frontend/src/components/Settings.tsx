import { useState, useEffect, useCallback } from "react";
import { C, MONO } from "../styles/colors";
import {
  GetConfig,
  UpdateConfig,
  GetStats,
  GetDBPath,
  GetVersion,
  SyncSetup,
  SyncNow,
  SyncStatus,
  SyncDisconnect,
  IsSyncConfigured,
  GetLaunchAtLogin,
  SetLaunchAtLogin,
} from "../bindings/snippetservice";
import type { SyncStatusInfo } from "../state/types";
import { Browser } from "@wailsio/runtime";

interface Config {
  Editor: string;
  DefaultLanguage: string;
  Theme: string;
  Find: {
    Sort: string;
    CopyToClipboard: boolean;
  };
}

interface Stats {
  TotalSnippets: number;
  TotalTags: number;
  Languages: Record<string, number>;
}

interface SettingsProps {
  onSortChanged?: () => void;
  onClose?: () => void;
}

export function Settings({ onSortChanged, onClose }: SettingsProps) {
  const [config, setConfig] = useState<Config | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [dbPath, setDbPath] = useState("");
  const [version, setVersion] = useState("");
  const [syncConfigured, setSyncConfigured] = useState(false);
  const [syncStatus, setSyncStatus] = useState<SyncStatusInfo | null>(null);
  const [showSyncModal, setShowSyncModal] = useState(false);
  const [token, setToken] = useState("");
  const [settingUp, setSettingUp] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [syncError, setSyncError] = useState<string | null>(null);
  const [launchAtLogin, setLaunchAtLogin] = useState(false);

  const loadConfig = useCallback(async () => {
    try {
      const cfg = await GetConfig();
      setConfig(cfg);
    } catch (err) {
      console.error("Failed to load config:", err);
    }
  }, []);

  useEffect(() => {
    GetLaunchAtLogin().then(setLaunchAtLogin).catch(console.error);
  }, []);

  useEffect(() => {
    loadConfig();
    GetStats()
      .then((s: any) => setStats(s))
      .catch(console.error);
    GetDBPath()
      .then(setDbPath)
      .catch(console.error);
    GetVersion()
      .then(setVersion)
      .catch(console.error);
  }, [loadConfig]);

  useEffect(() => {
    IsSyncConfigured().then((configured) => {
      setSyncConfigured(configured);
      if (configured) {
        SyncStatus().then(setSyncStatus).catch(console.error);
      }
    });
  }, []);

  async function updateField(updater: (cfg: Config) => Config) {
    if (!config) return;
    const updated = updater({ ...config });
    try {
      await UpdateConfig(updated as never);
      await loadConfig();
    } catch (err) {
      console.error("Failed to update config:", err);
    }
  }

  const handleSetup = async () => {
    setSettingUp(true);
    setSyncError(null);
    try {
      await SyncSetup(token);
      setShowSyncModal(false);
      setSyncConfigured(true);
      setToken("");
      const status = await SyncStatus();
      setSyncStatus(status);
    } catch (err: any) {
      setSyncError(err?.message || String(err) || "Failed to connect");
    } finally {
      setSettingUp(false);
    }
  };

  const handleSyncNow = async () => {
    setSyncing(true);
    try {
      await SyncNow();
      const status = await SyncStatus();
      setSyncStatus(status);
    } catch (err) {
      console.error("Sync failed:", err);
    } finally {
      setSyncing(false);
    }
  };

  const handleDisconnect = async () => {
    try {
      await SyncDisconnect();
      setSyncConfigured(false);
      setSyncStatus(null);
    } catch (err) {
      console.error("Disconnect failed:", err);
    }
  };

  if (!config) return null;

  const langCount = stats ? Object.keys(stats.Languages ?? {}).length : 0;

  return (
    <div
      style={{
        flex: 1,
        paddingTop: 40,
        overflow: "auto",
        background: C.bg,
      }}
    >
      <div style={{ padding: 24 }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            margin: "0 0 24px",
          }}
        >
          <h2
            style={{
              fontFamily: MONO,
              fontSize: 16,
              fontWeight: 600,
              color: C.text,
              margin: 0,
            }}
          >
            Settings
          </h2>
          {onClose && (
            <button
              onClick={onClose}
              style={{
                background: "transparent",
                border: "none",
                color: C.textDim,
                fontFamily: MONO,
                fontSize: 18,
                cursor: "pointer",
                padding: "4px 8px",
                borderRadius: 4,
                lineHeight: 1,
              }}
              title="Close settings"
            >
              &times;
            </button>
          )}
        </div>

        {/* GENERAL */}
        <SectionLabel>General</SectionLabel>
        <Divider />
        <Row label="Editor" value={config.Editor || "system default"} muted />
        <Row label="Theme" value="Catppuccin Mocha" muted />
        <ToggleRow
          label="Launch at login"
          value={launchAtLogin}
          onChange={async (v) => {
            try {
              await SetLaunchAtLogin(v);
              setLaunchAtLogin(v);
            } catch (err) {
              console.error("Failed to set launch at login:", err);
            }
          }}
        />

        {/* FIND */}
        <SectionLabel>Find</SectionLabel>
        <Divider />
        <SelectRow
          label="Default sort"
          value={config.Find?.Sort || "recent"}
          options={[
            { value: "recent", label: "Most recent" },
            { value: "usage", label: "Most used" },
            { value: "alpha", label: "Alphabetical" },
          ]}
          onChange={(v) => {
            updateField((c) => ({
              ...c,
              Find: { ...c.Find, Sort: v },
            })).then(() => onSortChanged?.());
          }}
        />
        <ToggleRow
          label="Copy to clipboard"
          value={config.Find?.CopyToClipboard ?? true}
          onChange={(v) =>
            updateField((c) => ({
              ...c,
              Find: { ...c.Find, CopyToClipboard: v },
            }))
          }
        />

        {/* DATA */}
        <SectionLabel>Data</SectionLabel>
        <Divider />
        <Row label="Database path" value={dbPath || "~/.local/share/snipt/"} />
        <Row label="Snippets" value={String(stats?.TotalSnippets ?? 0)} />
        <Row label="Languages" value={String(langCount)} />
        <Row label="Tags" value={String(stats?.TotalTags ?? 0)} />
        {!syncConfigured && (
          <div style={{ padding: "10px 0" }}>
            <Row label="Gist Sync" value="Not configured" muted />
            <button
              onClick={() => setShowSyncModal(true)}
              style={{
                background: "transparent",
                color: C.mauve,
                fontFamily: MONO,
                fontSize: 12,
                border: `1px solid ${C.border}`,
                borderRadius: 6,
                padding: "6px 14px",
                cursor: "pointer",
                marginTop: 4,
              }}
            >
              Set up sync
            </button>
          </div>
        )}
        {syncConfigured && syncStatus && (
          <>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 0", fontFamily: MONO, fontSize: 13 }}>
              <span style={{ color: C.textSub }}>Gist Sync</span>
              <span style={{ color: C.green }}>● Connected</span>
            </div>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 0", fontFamily: MONO, fontSize: 13 }}>
              <span style={{ color: C.textSub }}>Gist</span>
              <span
                style={{ color: C.mauve, cursor: "pointer", fontSize: 13 }}
                onClick={() => Browser.OpenURL(syncStatus.gist_url)}
              >
                {syncStatus.gist_id?.slice(0, 12)}...
              </span>
            </div>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 0", fontFamily: MONO, fontSize: 13 }}>
              <span style={{ color: C.textSub }}>Last sync</span>
              <span style={{ color: C.textDim }}>{timeAgo(syncStatus.last_sync)}</span>
            </div>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 0" }}>
              <span />
              <div style={{ display: "flex", gap: 8 }}>
                <button
                  onClick={handleSyncNow}
                  disabled={syncing}
                  style={{
                    background: "transparent",
                    color: C.mauve,
                    fontFamily: MONO,
                    fontSize: 12,
                    border: `1px solid ${C.border}`,
                    borderRadius: 6,
                    padding: "6px 14px",
                    cursor: syncing ? "not-allowed" : "pointer",
                    opacity: syncing ? 0.5 : 1,
                  }}
                >
                  {syncing ? "Syncing..." : "Sync now"}
                </button>
                <button
                  onClick={handleDisconnect}
                  style={{
                    background: "transparent",
                    color: C.red,
                    fontFamily: MONO,
                    fontSize: 12,
                    border: `1px solid ${C.border}`,
                    borderRadius: 6,
                    padding: "6px 14px",
                    cursor: "pointer",
                  }}
                >
                  Disconnect
                </button>
              </div>
            </div>
          </>
        )}

        {/* ABOUT */}
        <SectionLabel>About</SectionLabel>
        <Divider />
        <Row label="" value={`snipt ${version || "dev"}`} />
        <Row label="" value="Built with Go, Bubbletea, Wails" muted />
        <div style={{ padding: "10px 0" }}>
          <span
            style={{
              fontFamily: MONO,
              fontSize: 13,
              color: C.mauve,
              cursor: "pointer",
            }}
            onClick={() => Browser.OpenURL("https://github.com/infktd/snipt")}
          >
            github.com/infktd/snipt
          </span>
        </div>
      </div>
      {showSyncModal && (
        <div className="sync-modal-overlay" onClick={() => setShowSyncModal(false)}>
          <div className="sync-modal" onClick={(e) => e.stopPropagation()}>
            <h3>Connect to GitHub</h3>
            <p>
              Create a personal access token with the <code>gist</code> scope
              to sync your snippets across machines.
            </p>
            <span
              className="link"
              onClick={() =>
                Browser.OpenURL(
                  "https://github.com/settings/tokens/new?scopes=gist&description=snipt-sync"
                )
              }
            >
              Create token on GitHub →
            </span>
            <input
              type="password"
              placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              autoFocus
            />
            {syncError && (
              <p style={{ color: C.red, fontSize: 12, margin: "0 0 12px" }}>
                {syncError}
              </p>
            )}
            <div className="btn-row">
              <button
                className="btn-cancel"
                onClick={() => setShowSyncModal(false)}
              >
                Cancel
              </button>
              <button
                className="btn-connect"
                onClick={handleSetup}
                disabled={!token || settingUp}
              >
                {settingUp ? "Connecting..." : "Connect"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// --- Sub-components ---

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div
      style={{
        fontFamily: MONO,
        fontSize: 11,
        fontWeight: 600,
        color: C.mauve,
        textTransform: "uppercase",
        letterSpacing: 1.5,
        margin: "24px 0 12px",
      }}
    >
      {children}
    </div>
  );
}

function Divider() {
  return (
    <hr
      style={{
        border: "none",
        borderTop: `1px solid ${C.borderSubtle}`,
        margin: "0 0 16px",
      }}
    />
  );
}

function Row({
  label,
  value,
  muted,
}: {
  label: string;
  value: string;
  muted?: boolean;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      {label && <span style={{ color: C.textSub }}>{label}</span>}
      <span style={{ color: muted ? C.textDim : C.text }}>{value}</span>
    </div>
  );
}

function ToggleRow({
  label,
  value,
  onChange,
}: {
  label: string;
  value: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      <span style={{ color: C.textSub }}>{label}</span>
      <div
        onClick={() => onChange(!value)}
        style={{
          width: 36,
          height: 20,
          borderRadius: 10,
          background: value ? C.mauve : C.bgSurface,
          border: `1px solid ${value ? C.mauve : C.border}`,
          cursor: "pointer",
          position: "relative",
          transition: "background 0.15s ease",
        }}
      >
        <div
          style={{
            width: 14,
            height: 14,
            borderRadius: "50%",
            background: value ? C.bg : C.text,
            position: "absolute",
            top: 2,
            left: value ? 20 : 2,
            transition: "left 0.15s ease",
          }}
        />
      </div>
    </div>
  );
}

function SelectRow({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      <span style={{ color: C.textSub }}>{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          background: C.bgSurface,
          border: `1px solid ${C.border}`,
          borderRadius: 6,
          color: C.text,
          fontFamily: MONO,
          fontSize: 12,
          padding: "4px 10px",
          cursor: "pointer",
          outline: "none",
        }}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}

function timeAgo(iso: string): string {
  if (!iso) return "never";
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}
