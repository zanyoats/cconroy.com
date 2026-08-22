const root = document.documentElement;
const colorScheme = matchMedia("(prefers-color-scheme: dark)");

export default function registerTheme(up) {
    up.compiler(".top-nav", themeCompilerHandler);
}

function themeCompilerHandler(topNavElem) {
    const toggle = up.element.get(topNavElem, "[data-theme-toggle]");
    const icon = up.element.get(topNavElem, "[data-theme-icon]");

    if (!toggle || !icon) return;

    function updateButton() {
        const darkModeIsActive = currentTheme() === "dark";
        const nextTheme = darkModeIsActive ? "light" : "dark";

        toggle.setAttribute("aria-pressed", String(darkModeIsActive));
        toggle.setAttribute("aria-label", `Switch to ${nextTheme} theme`);

        icon.setAttribute(
            "href",
            darkModeIsActive ? "#icon-sun" : "#icon-moon",
        );
    }

    const unbindToggle = up.on(toggle, "click", () => {
        const nextTheme = currentTheme() === "dark" ? "light" : "dark";

        root.dataset.theme = nextTheme;
        saveTheme(nextTheme);
        updateButton();
    });

    const unbindColorScheme = up.on(colorScheme, "change", () => {
        // Continue following the OS until the user chooses an override.
        if (!root.dataset.theme) {
            updateButton();
        }
    });

    updateButton();

    return [unbindToggle, unbindColorScheme];
}

function systemTheme() {
    return colorScheme.matches ? "dark" : "light";
}

function currentTheme() {
    return root.dataset.theme || systemTheme();
}

function saveTheme(theme) {
    try {
        localStorage.setItem("theme", theme);
    } catch {
        // The theme still changes for this page, but cannot be persisted.
    }
}
