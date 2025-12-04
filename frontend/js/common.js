document.addEventListener("DOMContentLoaded", function() {
    const logo = document.getElementById("navLogo");
    if (!logo) return;

    // Get last logo from localStorage
    let lastLogo = localStorage.getItem("lastLogo") || "circle";

    // Determine the new logo src
    const newLogoSrc = lastLogo === "circle" ? "img/logo.png" : "img/logo-circle.png";
    const nextLogoState = lastLogo === "circle" ? "normal" : "circle";

    // Fade out, change src, fade in
    logo.style.opacity = 0; // start fade out
    setTimeout(() => {
        logo.src = newLogoSrc;
        logo.style.opacity = 1; // fade in
        localStorage.setItem("lastLogo", nextLogoState);
    }, 100); // half of the transition duration for smooth effect
});
