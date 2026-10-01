import { IconPlaceholder4 } from './IconPlaceholder4.jsx';

// figma node: 1180:35452 Button (32 variants)
const __venc = (v) => String(v).replace(/[%|=]/g, encodeURIComponent);
const __vkey = (p) => "variant=" + __venc(p.variant) + '|' + "state=" + __venc(p.state) + '|' + "size=" + __venc(p.size);

export function Button(_p = {}) {
  const props = { ..._p, label: _p.label ?? "Label", iconLeading: _p.iconLeading ?? false, iconTrailing: _p.iconTrailing ?? false, variant: _p.variant ?? "filled", state: _p.state ?? "resting", size: _p.size ?? "sm" };
  const __body0 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(47,112,239)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body1 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(22,85,209)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body2 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(0,51,160)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body3 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(31,39,46,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-regular-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body4 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(214,227,254)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body5 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(188,206,243)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body6 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(160,185,238)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body7 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(31,39,46,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-regular-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body8 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body9 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(223,228,238)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body10 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(208,216,230)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body11 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(213,214,215), 0 0 0 0.500px rgb(213,214,215)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-regular-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body12 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body13 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(223,228,238)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body14 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(208,216,230)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-default-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body15 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-default-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--spacing-8) * 1px)",
        padding: "12px 24px 12px 24px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-default-padding) * 1px)",
        paddingTop: "calc(var(--spacing-12) * 1px)",
        paddingRight: "calc(var(--button-default-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-12) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-regular-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "24px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 24,
            flexShrink: 0,
            alignSelf: "stretch",
            height: "auto",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 />}</div>
        )}
      </div>
    </div>
  );
  const __body16 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(47,112,239)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body17 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(22,85,209)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body18 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(0,51,160)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "var(--colour-scheme-primary-on-primary)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(255,255,255)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body19 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(31,39,46,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body20 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(214,227,254)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body21 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(188,206,243)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body22 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(160,185,238)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(0,51,160)",
          flexShrink: 0,
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(0,51,160)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body23 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(31,39,46,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
        }}>{props.text1 ?? "Label"}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body24 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body25 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(91,106,163,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body26 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(47,112,239), 0 0 0 0.500px rgb(47,112,239)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(208,216,230)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body27 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      boxShadow: "inset 0 0 0 0.500px rgb(213,214,215), 0 0 0 0.500px rgb(213,214,215)",
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body28 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body29 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgba(91,106,163,0.12)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body30 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        backgroundColor: "rgb(208,216,230)",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(47,112,239)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(47,112,239)",
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __body31 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      height: "calc(var(--button-small-height) * 1px)",
      overflow: "hidden",
      borderRadius: 100,
      display: "flex",
      flexDirection: "column",
      gap: 8,
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "row",
        gap: "calc(var(--button-small-gap) * 1px)",
        padding: "4px 16px 4px 16px",
        justifyContent: "center",
        alignItems: "center",
        flexWrap: "nowrap",
        boxSizing: "border-box",
        paddingLeft: "calc(var(--button-small-padding) * 1px)",
        paddingTop: "calc(var(--spacing-4-2) * 1px)",
        paddingRight: "calc(var(--button-small-padding) * 1px)",
        paddingBottom: "calc(var(--spacing-4-2) * 1px)",
        flexGrow: 1,
        alignSelf: "stretch",
      }}>
        {props.iconLeading && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
            color: "rgb(31,39,46)",
          }}>{props.iconLeadingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
        <span style={{
          position: "relative",
          opacity: 0.4,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 700,
          fontSize: "calc(var(--button-small-label-font-size) * 1px)",
          textAlign: "center",
          whiteSpace: "nowrap",
          lineHeight: "20px",
          letterSpacing: "0.100px",
          color: "rgb(31,39,46)",
          flexShrink: 0,
        }}>{props.label}</span>
        {props.iconTrailing && (
        <div style={{
            position: "relative",
            width: 16,
            height: 16,
            flexShrink: 0,
          }}>{props.iconTrailingSwap ?? <IconPlaceholder4 style={{ transform: "scale(0.667, 0.667)", transformOrigin: "0 0" }} />}</div>
        )}
      </div>
    </div>
  );
  const __impls = {
    // figma: Variant=filled, State=resting, Size=medium
    "variant=filled|state=resting|size=md": __body0,
    // figma: Variant=filled, State=hover, Size=medium
    "variant=filled|state=hover|size=md": __body1,
    // figma: Variant=filled, State=active, Size=medium
    "variant=filled|state=active|size=md": __body2,
    // figma: Variant=filled, State=disabled, Size=medium
    "variant=filled|state=disabled|size=md": __body3,
    // figma: Variant=tonal, State=resting, Size=medium
    "variant=tonal|state=resting|size=md": __body4,
    // figma: Variant=tonal, State=hover, Size=medium
    "variant=tonal|state=hover|size=md": __body5,
    // figma: Variant=tonal, State=active, Size=medium
    "variant=tonal|state=active|size=md": __body6,
    // figma: Variant=tonal, State=disabled, Size=medium
    "variant=tonal|state=disabled|size=md": __body7,
    // figma: Variant=outline, State=resting, Size=medium
    "variant=outline|state=resting|size=md": __body8,
    // figma: Variant=outline, State=hover, Size=medium
    "variant=outline|state=hover|size=md": __body9,
    // figma: Variant=outline, State=active, Size=medium
    "variant=outline|state=active|size=md": __body10,
    // figma: Variant=outline, State=disabled, Size=medium
    "variant=outline|state=disabled|size=md": __body11,
    // figma: Variant=text, State=resting, Size=medium
    "variant=text|state=resting|size=md": __body12,
    // figma: Variant=text, State=hover, Size=medium
    "variant=text|state=hover|size=md": __body13,
    // figma: Variant=text, State=active, Size=medium
    "variant=text|state=active|size=md": __body14,
    // figma: Variant=text, State=disabled, Size=medium
    "variant=text|state=disabled|size=md": __body15,
    // figma: Variant=filled, State=resting, Size=small
    "variant=filled|state=resting|size=sm": __body16,
    // figma: Variant=filled, State=hover, Size=small
    "variant=filled|state=hover|size=sm": __body17,
    // figma: Variant=filled, State=active, Size=small
    "variant=filled|state=active|size=sm": __body18,
    // figma: Variant=filled, State=disabled, Size=small
    "variant=filled|state=disabled|size=sm": __body19,
    // figma: Variant=tonal, State=resting, Size=small
    "variant=tonal|state=resting|size=sm": __body20,
    // figma: Variant=tonal, State=hover, Size=small
    "variant=tonal|state=hover|size=sm": __body21,
    // figma: Variant=tonal, State=active, Size=small
    "variant=tonal|state=active|size=sm": __body22,
    // figma: Variant=tonal, State=disabled, Size=small
    "variant=tonal|state=disabled|size=sm": __body23,
    // figma: Variant=outline, State=resting, Size=small
    "variant=outline|state=resting|size=sm": __body24,
    // figma: Variant=outline, State=hover, Size=small
    "variant=outline|state=hover|size=sm": __body25,
    // figma: Variant=outline, State=active, Size=small
    "variant=outline|state=active|size=sm": __body26,
    // figma: Variant=outline, State=disabled, Size=small
    "variant=outline|state=disabled|size=sm": __body27,
    // figma: Variant=text, State=resting, Size=small
    "variant=text|state=resting|size=sm": __body28,
    // figma: Variant=text, State=hover, Size=small
    "variant=text|state=hover|size=sm": __body29,
    // figma: Variant=text, State=active, Size=small
    "variant=text|state=active|size=sm": __body30,
    // figma: Variant=text, State=disabled, Size=small
    "variant=text|state=disabled|size=sm": __body31,
  };
  return (__impls[__vkey(props)] ?? __body16)();
}
export default Button;
