function isAuthenticated() {
    const token =
        localStorage.getItem("auth_token") ||
        sessionStorage.getItem("auth_token");

    const user =
        localStorage.getItem("auth_user") ||
        sessionStorage.getItem("auth_user");

    return !!(token && user);
}
