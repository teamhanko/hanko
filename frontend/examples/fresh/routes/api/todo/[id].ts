import { HandlerContext, Handlers } from "$fresh/server.ts";

export const handler: Handlers = {
  GET(_req: Request, _ctx: HandlerContext) {
    return new Response("Hello World");
  },
  PUT(_req: Request, _ctx: HandlerContext): Response {
    return new Response(null, {status: 200});
  },
  DELETE(_req: Request, _ctx: HandlerContext): Response {
    return new Response(null, {status: 200});
  },
};
