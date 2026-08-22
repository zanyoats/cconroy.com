import { up } from "./unpoly.js";
import registerTheme from "./theme.js";

// ensures browser Back/Forward restores only your main content
up.history.config.restoreTargets = ["[up-main]"];

registerTheme(up);

if (document.readyState === "complete") {
    up.boot();
} else {
    window.addEventListener("load", () => up.boot(), { once: true });
}
