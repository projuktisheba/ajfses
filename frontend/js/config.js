(function () {
    const hostname = window.location.hostname;

    let API_URL = "";
    console.log("Host: ", hostname, "type: ", typeof hostname)
    if (hostname == "" || hostname === "localhost" || hostname === "127.0.0.1") {
        API_URL = "http://localhost:8080/api/v1";    // dev URL
    } else {
        API_URL = "https://ajfses-api.pssoft.xyz/api/v1";  // production URL
    }
    window.env = { API_URL };
})();
