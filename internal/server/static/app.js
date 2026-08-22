import { up } from "./unpoly.js";
import registerTheme from "./theme.js";

// ensures browser Back/Forward restores only your main content
up.history.config.restoreTargets = ["[up-main]"];

registerTheme(up);
