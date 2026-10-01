function syncActiveNav() {
    var path = window.location.pathname;
    document.querySelectorAll(".sidebar-item").forEach(function (item) {
        var link = item.querySelector("a");
        var href = link.getAttribute("href");
        var isRoot = href === "/" && path === "/";
        var isSection = href !== "/" && (path === href || path.startsWith(href + "/"));
        item.classList.toggle("active", isRoot || isSection);
    });
}

document.addEventListener("htmx:afterSettle", syncActiveNav);
document.addEventListener("DOMContentLoaded", syncActiveNav);
