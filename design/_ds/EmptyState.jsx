import { FilledButton3 } from './FilledButton3.jsx';
import { OutlinedButton2 } from './OutlinedButton2.jsx';
import { Placeholder } from './Placeholder.jsx';
import { TextButton3 } from './TextButton3.jsx';

// figma node: 1:8662 Empty state (2 variants)
const __venc = (v) => String(v).replace(/[%|=]/g, encodeURIComponent);
const __vkey = (p) => "type=" + __venc(p.type);

export function EmptyState(_p = {}) {
  const props = { ..._p, type: _p.type ?? "large container", heading: _p.heading ?? "Empty state heading", description: _p.description ?? "The description can run across multiple lines and can use bold text.", actions: _p.actions ?? true, secondaryButton: _p.secondaryButton ?? true };
  const __body0 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      borderRadius: 16,
      display: "flex",
      flexDirection: "column",
      gap: "calc(var(--spacing-24) * 1px)",
      padding: "24px 32px 24px 32px",
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      boxSizing: "border-box",
      paddingLeft: "calc(var(--spacing-32) * 1px)",
      paddingTop: "calc(var(--spacing-24) * 1px)",
      paddingRight: "calc(var(--spacing-32) * 1px)",
      paddingBottom: "calc(var(--spacing-24) * 1px)",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
          position: "relative",
          height: 170,
          flexShrink: 0,
          alignSelf: "stretch",
          width: "auto",
        }}>{props.illustration ?? <Placeholder />}</div>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "column",
        gap: "calc(var(--spacing-12) * 1px)",
        alignItems: "center",
        flexWrap: "nowrap",
        flexShrink: 0,
        alignSelf: "stretch",
      }}>
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 600,
          fontSize: 32,
          textAlign: "center",
          lineHeight: "40px",
          color: "var(--colour-scheme-surface-on-surface)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.heading}</span>
        <span style={{
          position: "relative",
          maxWidth: 420,
          maxHeight: null,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 400,
          fontSize: 16,
          textAlign: "center",
          lineHeight: "28px",
          color: "var(--colour-scheme-surface-on-surface)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.description}</span>
      </div>
      {props.actions && (
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "column",
        gap: "calc(var(--spacing-12) * 1px)",
        alignItems: "center",
        flexWrap: "nowrap",
        flexShrink: 0,
        alignSelf: "stretch",
      }}>
        <FilledButton3
          style={{
            position: "relative",
            height: 48,
            width: 180,
            flexShrink: 0,
          }}
          labelText={"Primary action"}
          state={"enabled"}
          size={"default"}
          iconSlot={"leading"}
          showIcon={false}
        />
        {props.secondaryButton && (
        <TextButton3
          style={{ position: "relative", height: 48, flexShrink: 0 }}
          labelText={"Action"}
          state={"enabled"}
          size={"default"}
          iconSlot={"leading"}
          showIcon={false}
        />
        )}
      </div>
      )}
    </div>
  );
  const __body1 = () => (
    <div className={props.className} style={{
      width: "fit-content",
      borderRadius: 16,
      display: "flex",
      flexDirection: "column",
      gap: "calc(var(--spacing-16) * 1px)",
      padding: "16px 16px 16px 16px",
      justifyContent: "center",
      alignItems: "center",
      flexWrap: "nowrap",
      boxSizing: "border-box",
      paddingLeft: "calc(var(--spacing-16) * 1px)",
      paddingTop: "calc(var(--spacing-16) * 1px)",
      paddingRight: "calc(var(--spacing-16) * 1px)",
      paddingBottom: "calc(var(--spacing-16) * 1px)",
      position: "relative",
      ...props.style,
    }}>
      <div style={{
          position: "relative",
          height: 136,
          flexShrink: 0,
          alignSelf: "stretch",
          width: "auto",
        }}>{props.illustration ?? <Placeholder />}</div>
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "column",
        gap: "calc(var(--spacing-8) * 1px)",
        alignItems: "center",
        flexWrap: "nowrap",
        flexShrink: 0,
        alignSelf: "stretch",
      }}>
        <span style={{
          position: "relative",
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 600,
          fontSize: 24,
          textAlign: "center",
          lineHeight: "36px",
          color: "var(--colour-scheme-surface-on-surface)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.heading}</span>
        <span style={{
          position: "relative",
          maxWidth: 420,
          maxHeight: null,
          fontFamily: "\"Nunito Sans\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", Arial, sans-serif",
          fontWeight: 400,
          fontSize: 16,
          textAlign: "center",
          lineHeight: "28px",
          color: "var(--colour-scheme-surface-on-surface)",
          flexShrink: 0,
          alignSelf: "stretch",
        }}>{props.description}</span>
      </div>
      {props.actions && (
      <div style={{
        position: "relative",
        display: "flex",
        flexDirection: "column",
        gap: "calc(var(--spacing-12) * 1px)",
        alignItems: "center",
        flexWrap: "nowrap",
        flexShrink: 0,
        alignSelf: "stretch",
      }}>
        <OutlinedButton2
          style={{
            position: "relative",
            height: 48,
            width: 180,
            flexShrink: 0,
          }}
          labelText={"Primary action"}
          state={"enabled"}
          size={"default"}
          iconSlot={"leading"}
          showIcon={false}
        />
        {props.secondaryButton && (
        <TextButton3
          style={{ position: "relative", height: 48, flexShrink: 0 }}
          labelText={"Action"}
          state={"enabled"}
          size={"default"}
          iconSlot={"leading"}
          showIcon={false}
        />
        )}
      </div>
      )}
    </div>
  );
  const __impls = {
    // figma: Type=Full page
    "type=full page": __body0,
    // figma: Type=Large container
    "type=large container": __body1,
  };
  return (__impls[__vkey(props)] ?? __body1)();
}
export default EmptyState;
