window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    urls: [
      { name: "Orders API", url: "./orders.swagger.json" },
      { name: "Admin API",  url: "./admin.swagger.json" }
    ],
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    requestInterceptor: (request) => {
      if (request.url.includes('/v1/')) {
        const currentHost = window.location.hostname;
        let path = request.url;
        if (path.includes('://')) {
          const urlObj = new URL(path);
          path = urlObj.pathname + urlObj.search;
        }
        request.url = `http://${currentHost}:8080${path}`;
      }
      return request;
    }
  });

  //</editor-fold>
};
