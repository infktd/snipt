import { useEffect, useState } from "react";
import { C, MONO, BODY } from "../styles/colors";
import { GetVersion } from "../bindings/snippetservice";
import { Events } from "@wailsio/runtime";

export function AboutDialog() {
  const [open, setOpen] = useState(false);
  const [version, setVersion] = useState("");

  useEffect(() => {
    const cancel = Events.On("open-about", () => {
      GetVersion()
        .then(setVersion)
        .catch(() => setVersion("dev"));
      setOpen(true);
    });
    return cancel;
  }, []);

  if (!open) return null;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "rgba(0, 0, 0, 0.6)",
        zIndex: 200,
      }}
      onClick={() => setOpen(false)}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: C.bgCard,
          border: `1px solid ${C.border}`,
          borderRadius: 12,
          padding: "32px 40px",
          minWidth: 300,
          fontFamily: BODY,
          textAlign: "center",
        }}
      >
        <div
          style={{
            display: "inline-block",
            background: `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
            color: C.bg,
            padding: "6px 14px",
            borderRadius: 8,
            fontFamily: MONO,
            fontSize: 16,
            fontWeight: 700,
            letterSpacing: 0.5,
            marginBottom: 16,
          }}
        >
          snipt &lt;/&gt;
        </div>

        <div
          style={{
            color: C.textDim,
            fontFamily: MONO,
            fontSize: 12,
            marginBottom: 12,
          }}
        >
          Version {version || "dev"}
        </div>

        <div
          style={{
            color: C.textSub,
            fontSize: 13,
            lineHeight: 1.5,
            marginBottom: 24,
          }}
        >
          A snippet manager for the
          <br />
          command line and beyond.
        </div>

        <button
          onClick={() => setOpen(false)}
          style={{
            padding: "8px 24px",
            borderRadius: 6,
            border: `1px solid ${C.border}`,
            background: C.bgSurface,
            color: C.textSub,
            fontFamily: BODY,
            fontSize: 13,
            cursor: "pointer",
          }}
        >
          OK
        </button>
      </div>
    </div>
  );
}
