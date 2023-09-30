import { AppProps } from "$fresh/server.ts";

export default function App({ Component }: AppProps) {
  return (
    <html>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Hanko App</title>
        <link rel="stylesheet" href="/app.css"/>
      </head>
      <body>
      <div class="p-4 mx-auto max-w-screen-md">
        <Component />
        </div>
      </body>
    </html>
  );
}
