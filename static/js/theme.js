const COMPANIES = {
    default: { name: "חברת הביטוח", initial: "ב" },
    phoenix: { name: "הפניקס", initial: "פ" },
    hapoel: { name: "הפול חברת ביטוח", initial: "ה" },
    agri: { name: "ביטוח חקלאי", initial: "ח" },
    shmeret: { name: "שמרת ביטוח", initial: "ש" },
};

function applyTheme() {
    var company = localStorage.getItem("insurance-company") || "default";
    var mode = localStorage.getItem("insurance-theme") || "light";

    document.documentElement.setAttribute("data-company", company);
    document.documentElement.setAttribute("data-theme", mode);

    var meta = COMPANIES[company] || COMPANIES.default;
    var icon = document.getElementById("workspace-icon");
    var name = document.getElementById("workspace-name");
    if (icon) icon.textContent = meta.initial;
    if (name) name.textContent = meta.name;

    var select = document.getElementById("company-switcher");
    if (select) select.value = company;

    var toggleIcon = document.getElementById("theme-toggle-icon");
    var toggleLabel = document.getElementById("theme-toggle-label");
    if (toggleIcon && toggleLabel) {
        if (mode === "dark") {
            toggleIcon.className = "bi bi-sun-fill";
            toggleLabel.textContent = "מצב בהיר";
        } else {
            toggleIcon.className = "bi bi-moon-stars-fill";
            toggleLabel.textContent = "מצב כהה";
        }
    }
}

function bindThemeControls() {
    var select = document.getElementById("company-switcher");
    if (select) {
        select.addEventListener("change", function () {
            localStorage.setItem("insurance-company", select.value);
            applyTheme();
        });
    }

    var toggle = document.getElementById("theme-toggle");
    if (toggle) {
        toggle.addEventListener("click", function () {
            var current = localStorage.getItem("insurance-theme") || "light";
            localStorage.setItem("insurance-theme", current === "light" ? "dark" : "light");
            applyTheme();
        });
    }
}

document.addEventListener("DOMContentLoaded", function () {
    applyTheme();
    bindThemeControls();
});
