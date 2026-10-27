// The fork's update service reuses upstream's release selection and ledger.
// Only update metadata and download redirects are exposed on this domain.
import worker from "./worker.js";

export default {
  async fetch(req, env, ctx) {
    if (req.method !== "GET" && req.method !== "HEAD") {
      return new Response("Method not allowed\n", { status: 405, headers: { Allow: "GET, HEAD" } });
    }
    const path = new URL(req.url).pathname;
    if (path === "/") {
      return Response.json({
        service: "magpie updates",
        repository: "zhumengstarsea/magpie",
        latest: "/api/latest",
        stable: "/api/latest?channel=stable",
        notes: "/api/notes",
      });
    }
    if (path === "/api/latest" || path === "/api/notes" || path === "/download" || path.startsWith("/download/")) {
      const res = await worker.fetch(req, env, ctx);
      return req.method === "HEAD" ? new Response(null, res) : res;
    }
    return new Response("Not found\n", { status: 404 });
  },
  scheduled: worker.scheduled,
};
